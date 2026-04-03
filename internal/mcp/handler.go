package mcp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/djtouchette/vaulty/internal/audit"
	"github.com/djtouchette/vaulty/internal/executor"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/djtouchette/vaulty/internal/proxy"
	"github.com/djtouchette/vaulty/internal/vault"
)

// Handler processes MCP tool calls using the vault, config, and redactor.
type Handler struct {
	vault     *vault.Vault
	config    *policy.Config
	redactor  *proxy.Redactor
	logger    *audit.Logger
	approvals *ApprovalStore
}

// NewHandler creates a new MCP handler.
func NewHandler(v *vault.Vault, cfg *policy.Config, logger *audit.Logger) *Handler {
	secrets := make(map[string]string)
	for _, name := range v.List() {
		val, _ := v.Get(name)
		secrets[name] = val
	}

	return &Handler{
		vault:     v,
		config:    cfg,
		redactor:  proxy.NewRedactor(secrets),
		logger:    logger,
		approvals: NewApprovalStore(5 * time.Minute),
	}
}

// HandleToolCall dispatches a tool call by name.
func (h *Handler) HandleToolCall(name string, args json.RawMessage) (string, error) {
	switch name {
	case "vaulty_request":
		return h.handleRequest(args)
	case "vaulty_exec":
		return h.handleExec(args)
	case "vaulty_list":
		return h.handleList()
	case "vaulty_approve":
		return h.handleApprove(args)
	case "vaulty_pending":
		return h.handleListPending()
	case "vaulty_list_services":
		return h.handleListServices()
	case "vaulty_check_access":
		return h.handleCheckAccess(args)
	case "vaulty_secret_metadata":
		return h.handleSecretMetadata()
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (h *Handler) handleRequest(args json.RawMessage) (string, error) {
	var params struct {
		Method     string            `json:"method"`
		URL        string            `json:"url"`
		SecretName string            `json:"secret_name"`
		Headers    map[string]string `json:"headers"`
		Body       string            `json:"body"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Policy check
	if err := h.config.ValidateDomain(params.SecretName, params.URL); err != nil {
		h.logger.LogDenied(params.SecretName, params.URL, err.Error())
		return "", err
	}

	// Check if approval is needed
	sp := h.config.GetSecretPolicy(params.SecretName)
	if !sp.AutoApprove {
		pa := h.approvals.Create(params.SecretName, params.URL, "proxy", args)
		return fmt.Sprintf("Approval required: secret %q will be used for %s %s.\nApproval ID: %s\nCall vaulty_approve with {\"approval_id\": %q, \"decision\": \"approve\"} to proceed, or \"deny\" to reject.",
			params.SecretName, params.Method, params.URL, pa.ID, pa.ID), nil
	}

	return h.executeRequest(args)
}

func (h *Handler) executeRequest(args json.RawMessage) (string, error) {
	var params struct {
		Method     string            `json:"method"`
		URL        string            `json:"url"`
		SecretName string            `json:"secret_name"`
		Headers    map[string]string `json:"headers"`
		Body       string            `json:"body"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	secretValue, ok := h.vault.Get(params.SecretName)
	if !ok {
		return "", fmt.Errorf("secret %q not found", params.SecretName)
	}

	sp := h.config.GetSecretPolicy(params.SecretName)
	mode := proxy.InjectMode(sp.InjectAs)

	result, err := proxy.DoRequest(params.Method, params.URL, params.Headers, params.Body, secretValue, mode, sp.HeaderName, h.redactor)
	if err != nil {
		return "", err
	}

	h.logger.LogProxy(params.SecretName, params.Method, params.URL, result.StatusCode)

	// Format response for the agent
	var sb strings.Builder
	fmt.Fprintf(&sb, "HTTP %d\n", result.StatusCode)
	for k, v := range result.Headers {
		fmt.Fprintf(&sb, "%s: %s\n", k, v)
	}
	sb.WriteString("\n")
	sb.WriteString(result.Body)

	return sb.String(), nil
}

func (h *Handler) handleExec(args json.RawMessage) (string, error) {
	var params struct {
		Command    string   `json:"command"`
		Secrets    []string `json:"secrets"`
		WorkingDir string   `json:"working_dir"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Validate all secrets before creating approval
	for _, name := range params.Secrets {
		if err := h.config.ValidateCommand(name, params.Command); err != nil {
			h.logger.LogDenied(name, params.Command, err.Error())
			return "", err
		}
	}

	// Check if approval is needed for any secret
	needsApproval := false
	for _, name := range params.Secrets {
		sp := h.config.GetSecretPolicy(name)
		if !sp.AutoApprove {
			needsApproval = true
			break
		}
	}

	if needsApproval {
		secretNames := strings.Join(params.Secrets, ", ")
		pa := h.approvals.Create(secretNames, params.Command, "exec", args)
		return fmt.Sprintf("Approval required: secrets [%s] will be injected for command %q.\nApproval ID: %s\nCall vaulty_approve with {\"approval_id\": %q, \"decision\": \"approve\"} to proceed, or \"deny\" to reject.",
			secretNames, params.Command, pa.ID, pa.ID), nil
	}

	return h.executeExec(args)
}

func (h *Handler) executeExec(args json.RawMessage) (string, error) {
	var params struct {
		Command    string   `json:"command"`
		Secrets    []string `json:"secrets"`
		WorkingDir string   `json:"working_dir"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Resolve secrets
	secretMap := make(map[string]string)
	for _, name := range params.Secrets {
		val, ok := h.vault.Get(name)
		if !ok {
			return "", fmt.Errorf("secret %q not found", name)
		}
		secretMap[name] = val
	}

	result, err := executor.Run(params.Command, secretMap, params.WorkingDir, h.redactor)
	if err != nil {
		return "", err
	}

	for _, name := range params.Secrets {
		h.logger.LogExec(name, params.Command, result.ExitCode)
	}

	var sb strings.Builder
	if result.Stdout != "" {
		sb.WriteString(result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprintf(&sb, "\n[stderr]\n%s", result.Stderr)
	}
	fmt.Fprintf(&sb, "\n[exit code: %d]", result.ExitCode)

	return sb.String(), nil
}

func (h *Handler) handleList() (string, error) {
	names := h.vault.List()
	if len(names) == 0 {
		return "No secrets configured.", nil
	}

	var sb strings.Builder
	for _, name := range names {
		sp := h.config.GetSecretPolicy(name)
		fmt.Fprintf(&sb, "- %s", name)
		if sp.Description != "" {
			fmt.Fprintf(&sb, ": %s", sp.Description)
		}
		if len(sp.AllowedDomains) > 0 {
			fmt.Fprintf(&sb, " (domains: %s)", strings.Join(sp.AllowedDomains, ", "))
		}
		if len(sp.AllowedCommands) > 0 {
			fmt.Fprintf(&sb, " (commands: %s)", strings.Join(sp.AllowedCommands, ", "))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func (h *Handler) handleApprove(args json.RawMessage) (string, error) {
	var params struct {
		ApprovalID string `json:"approval_id"`
		Decision   string `json:"decision"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if params.ApprovalID == "" {
		return "", fmt.Errorf("approval_id is required")
	}
	if params.Decision != "approve" && params.Decision != "deny" {
		return "", fmt.Errorf("decision must be \"approve\" or \"deny\"")
	}

	pa, ok := h.approvals.Get(params.ApprovalID)
	if !ok {
		return "", fmt.Errorf("approval %q not found", params.ApprovalID)
	}

	if params.Decision == "deny" {
		if err := h.approvals.Deny(params.ApprovalID); err != nil {
			return "", err
		}
		h.logger.LogApproval(pa.SecretName, pa.Target, "denied")
		return fmt.Sprintf("Denied: %s access to %s via %s.", pa.SecretName, pa.Target, pa.Action), nil
	}

	if err := h.approvals.Approve(params.ApprovalID); err != nil {
		return "", err
	}
	h.logger.LogApproval(pa.SecretName, pa.Target, "approved")

	// Execute the original action
	switch pa.Action {
	case "proxy":
		return h.executeRequest(pa.Args)
	case "exec":
		return h.executeExec(pa.Args)
	default:
		return "", fmt.Errorf("unknown action: %s", pa.Action)
	}
}

func (h *Handler) handleListPending() (string, error) {
	h.approvals.Cleanup()
	pending := h.approvals.ListPending()

	if len(pending) == 0 {
		return "No pending approvals.", nil
	}

	var sb strings.Builder
	for _, pa := range pending {
		fmt.Fprintf(&sb, "- %s: %s %s (%s) — expires %s\n",
			pa.ID, pa.Action, pa.Target, pa.SecretName,
			pa.ExpiresAt.Format("15:04:05"))
	}
	return sb.String(), nil
}

// T-151: Discovery tools

func (h *Handler) handleListServices() (string, error) {
	if h.config.Secrets == nil || len(h.config.Secrets) == 0 {
		return "No services configured.", nil
	}

	var sb strings.Builder
	for name, sp := range h.config.Secrets {
		fmt.Fprintf(&sb, "- %s\n", name)
		if sp.Description != "" {
			fmt.Fprintf(&sb, "  Description: %s\n", sp.Description)
		}
		if len(sp.AllowedDomains) > 0 {
			fmt.Fprintf(&sb, "  Domains: %s\n", strings.Join(sp.AllowedDomains, ", "))
		}
		if sp.InjectAs != "" {
			fmt.Fprintf(&sb, "  Injection: %s\n", sp.InjectAs)
		}
		if sp.HeaderName != "" {
			fmt.Fprintf(&sb, "  Header: %s\n", sp.HeaderName)
		}
		if len(sp.AllowedCommands) > 0 {
			fmt.Fprintf(&sb, "  Commands: %s\n", strings.Join(sp.AllowedCommands, ", "))
		}
		if sp.AutoApprove {
			fmt.Fprintf(&sb, "  Auto-approve: yes\n")
		}
	}
	return sb.String(), nil
}

func (h *Handler) handleCheckAccess(args json.RawMessage) (string, error) {
	var params struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	if params.URL == "" {
		return "", fmt.Errorf("url is required")
	}

	parsed, err := url.Parse(params.URL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	host := parsed.Hostname()

	var matches []string
	for name, sp := range h.config.Secrets {
		if len(sp.AllowedDomains) == 0 {
			// Wildcard: no restrictions means it can access any domain
			matches = append(matches, name+" (wildcard — all domains)")
			continue
		}
		for _, domain := range sp.AllowedDomains {
			if strings.EqualFold(host, domain) {
				matches = append(matches, name)
				break
			}
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No secrets have access to %s. Check your vaulty.toml allowed_domains configuration.", host), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Secrets with access to %s:\n", host)
	for _, m := range matches {
		fmt.Fprintf(&sb, "- %s\n", m)
	}
	return sb.String(), nil
}

func (h *Handler) handleSecretMetadata() (string, error) {
	names := h.vault.List()
	if len(names) == 0 {
		return "No secrets stored.", nil
	}

	var sb strings.Builder
	for _, name := range names {
		sp := h.config.GetSecretPolicy(name)
		fmt.Fprintf(&sb, "- %s\n", name)
		if sp.Description != "" {
			fmt.Fprintf(&sb, "  Description: %s\n", sp.Description)
		}
		if len(sp.AllowedDomains) > 0 {
			fmt.Fprintf(&sb, "  Domains: %s\n", strings.Join(sp.AllowedDomains, ", "))
		}
		if len(sp.AllowedCommands) > 0 {
			fmt.Fprintf(&sb, "  Commands: %s\n", strings.Join(sp.AllowedCommands, ", "))
		}
		if sp.InjectAs != "" {
			fmt.Fprintf(&sb, "  Injection: %s\n", sp.InjectAs)
		}
		if sp.AutoApprove {
			fmt.Fprintf(&sb, "  Auto-approve: yes\n")
		}
	}
	return sb.String(), nil
}
