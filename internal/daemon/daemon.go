package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/djtouchette/vaulty/internal/audit"
	"github.com/djtouchette/vaulty/internal/executor"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/djtouchette/vaulty/internal/proxy"
	"github.com/djtouchette/vaulty/internal/vault"
)

// Daemon is the long-running process that holds secrets and serves requests.
type Daemon struct {
	vaults   map[string]*vault.Vault
	config   *policy.Config
	redactor *proxy.Redactor
	logger   *audit.Logger
	notifier *Notifier

	httpServer   *http.Server
	socketPath   string
	socketLn     net.Listener
	pidPath      string
	idleTimeout  time.Duration
	idleTimer    *time.Timer
	mu           sync.Mutex
}

// New creates a new daemon with the given vaults and config.
// The vaults map keys are vault names ("" for the default vault).
func New(vaults map[string]*vault.Vault, cfg *policy.Config) (*Daemon, error) {
	// Build combined redactor from all vaults
	secrets := make(map[string]string)
	for _, v := range vaults {
		for _, name := range v.List() {
			val, _ := v.Get(name)
			secrets[name] = val
		}
	}
	redactor := proxy.NewRedactor(secrets)

	logPath := "~/.config/vaulty/audit.log"
	auditLogger, err := audit.NewLogger(logPath)
	if err != nil {
		return nil, fmt.Errorf("creating audit logger: %w", err)
	}

	idleTimeout := 8 * time.Hour
	if cfg.Vault.IdleTimeout != "" {
		d, err := time.ParseDuration(cfg.Vault.IdleTimeout)
		if err == nil {
			idleTimeout = d
		}
	}

	return &Daemon{
		vaults:      vaults,
		config:      cfg,
		redactor:    redactor,
		logger:      auditLogger,
		notifier:    NewNotifier(cfg.Vault.Notifications),
		socketPath:  cfg.Vault.Socket,
		pidPath:     pidFilePath(),
		idleTimeout: idleTimeout,
	}, nil
}

// getVault returns the vault for the given name ("" for default).
func (d *Daemon) getVault(name string) (*vault.Vault, error) {
	v, ok := d.vaults[name]
	if !ok {
		if name == "" {
			return nil, fmt.Errorf("default vault not loaded")
		}
		return nil, fmt.Errorf("vault %q not loaded — start the daemon with --vaults %s", name, name)
	}
	return v, nil
}

// Run starts the daemon and blocks until shutdown.
func (d *Daemon) Run(ctx context.Context) error {
	// Write PID file
	if err := d.writePID(); err != nil {
		return err
	}
	defer d.cleanup()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/request", d.handleRequest)

	errCh := make(chan error, 2)

	// Start platform-specific socket listener (Unix socket on Unix, no-op on Windows)
	if d.socketPath != "" {
		ln, socketServer, err := StartSocketListener(d.socketPath, mux)
		if err != nil {
			return fmt.Errorf("listening on socket %s: %w", d.socketPath, err)
		}
		if ln != nil {
			d.socketLn = ln
			defer socketServer.Close()
		}
	}

	// Start HTTP listener (works on all platforms)
	if d.config.Vault.HTTPPort > 0 {
		addr := fmt.Sprintf("127.0.0.1:%d", d.config.Vault.HTTPPort)
		d.httpServer = &http.Server{
			Addr:    addr,
			Handler: mux,
		}
		go func() {
			log.Printf("Listening on http://%s", addr)
			if err := d.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}()
		defer d.httpServer.Close()
	}

	// Set up idle timeout
	d.idleTimer = time.NewTimer(d.idleTimeout)

	// Wait for shutdown signal (platform-specific)
	return WaitForSignal(ctx, d.idleTimer, errCh)
}

func (d *Daemon) cleanup() {
	for _, v := range d.vaults {
		v.Zero()
	}
	log.Println("Secrets cleared from memory.")

	if d.logger != nil {
		d.logger.Close()
	}

	os.Remove(d.pidPath)

	if d.socketPath != "" {
		os.Remove(d.socketPath)
	}
}

func (d *Daemon) resetIdle() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.idleTimer != nil {
		d.idleTimer.Reset(d.idleTimeout)
	}
}

func (d *Daemon) handleRequest(w http.ResponseWriter, r *http.Request) {
	d.resetIdle()

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Error: "POST required"})
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: "invalid request: " + err.Error()})
		return
	}

	var resp Response
	switch req.Action {
	case "proxy":
		resp = d.handleProxy(req)
	case "exec":
		resp = d.handleExec(req)
	case "list":
		resp = d.handleList()
	default:
		resp = Response{Error: fmt.Sprintf("unknown action: %q", req.Action)}
	}

	status := http.StatusOK
	if resp.Error != "" {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, resp)
}

func (d *Daemon) handleProxy(req Request) Response {
	if req.Secret == "" {
		return Response{Error: "secret name required"}
	}
	if req.URL == "" {
		return Response{Error: "url required"}
	}
	if req.Method == "" {
		req.Method = "GET"
	}

	// Policy check
	if err := d.config.ValidateDomain(req.Secret, req.URL); err != nil {
		d.logger.LogDenied(req.Secret, req.URL, err.Error())
		d.notifier.NotifyDenied(req.Secret, req.URL, err.Error())
		return Response{Error: err.Error()}
	}

	// Resolve vault: use request vault, fall back to policy vault, then default
	vaultName := req.Vault
	if vaultName == "" {
		sp := d.config.GetSecretPolicy(req.Secret)
		vaultName = sp.Vault
	}
	v, err := d.getVault(vaultName)
	if err != nil {
		return Response{Error: err.Error()}
	}

	// Get secret value
	secretValue, ok := v.Get(req.Secret)
	if !ok {
		return Response{Error: fmt.Sprintf("secret %q not found", req.Secret)}
	}

	// Get injection mode
	sp := d.config.GetSecretPolicy(req.Secret)
	mode := proxy.InjectMode(sp.InjectAs)

	result, err := proxy.DoRequest(req.Method, req.URL, req.Headers, req.Body, secretValue, mode, sp.HeaderName, d.redactor)
	if err != nil {
		return Response{Error: err.Error()}
	}

	d.logger.LogProxy(req.Secret, req.Method, req.URL, result.StatusCode)

	return Response{
		OK:      true,
		Status:  result.StatusCode,
		Headers: result.Headers,
		Body:    result.Body,
	}
}

func (d *Daemon) handleExec(req Request) Response {
	if req.Command == "" {
		return Response{Error: "command required"}
	}

	secretNames := req.Secrets
	if req.Secret != "" && len(secretNames) == 0 {
		secretNames = []string{req.Secret}
	}

	// Resolve secrets and validate policies
	secrets := make(map[string]string)
	for _, name := range secretNames {
		if err := d.config.ValidateCommand(name, req.Command); err != nil {
			d.logger.LogDenied(name, req.Command, err.Error())
			d.notifier.NotifyDenied(name, req.Command, err.Error())
			return Response{Error: err.Error()}
		}

		// Resolve vault: use request vault, fall back to policy vault, then default
		vaultName := req.Vault
		if vaultName == "" {
			sp := d.config.GetSecretPolicy(name)
			vaultName = sp.Vault
		}
		v, err := d.getVault(vaultName)
		if err != nil {
			return Response{Error: err.Error()}
		}

		val, ok := v.Get(name)
		if !ok {
			return Response{Error: fmt.Sprintf("secret %q not found", name)}
		}
		secrets[name] = val
	}

	result, err := executor.Run(req.Command, secrets, req.WorkDir, d.redactor)
	if err != nil {
		return Response{Error: err.Error()}
	}

	for _, name := range secretNames {
		d.logger.LogExec(name, req.Command, result.ExitCode)
	}

	return Response{
		OK:       true,
		ExitCode: &result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
	}
}

func (d *Daemon) handleList() Response {
	var infos []SecretInfo
	for vaultName, v := range d.vaults {
		for _, name := range v.List() {
			sp := d.config.GetSecretPolicy(name)
			info := SecretInfo{
				Name:            name,
				Description:     sp.Description,
				AllowedDomains:  sp.AllowedDomains,
				AllowedCommands: sp.AllowedCommands,
				InjectAs:        sp.InjectAs,
			}
			if vaultName != "" {
				info.Vault = vaultName
			}
			infos = append(infos, info)
		}
	}
	return Response{
		OK:         true,
		SecretList: infos,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (d *Daemon) writePID() error {
	dir := filepath.Dir(d.pidPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating pid directory: %w", err)
	}
	return os.WriteFile(d.pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0600)
}

func pidFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "vaulty", "vaulty.pid")
}

// PIDFilePath returns the path to the daemon PID file (for use by stop command).
func PIDFilePath() string {
	return pidFilePath()
}

// SocketPath returns the default socket path.
func SocketPath() string {
	return "/tmp/vaulty.sock"
}
