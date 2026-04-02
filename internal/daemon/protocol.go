package daemon

// Request is the JSON request format for daemon communication.
type Request struct {
	Action  string            `json:"action"`            // proxy, exec, list
	Method  string            `json:"method,omitempty"`  // HTTP method (for proxy)
	URL     string            `json:"url,omitempty"`     // target URL (for proxy)
	Secret  string            `json:"secret,omitempty"`  // secret name
	Secrets []string          `json:"secrets,omitempty"` // multiple secret names (for exec)
	Headers map[string]string `json:"headers,omitempty"` // extra headers (for proxy)
	Body    string            `json:"body,omitempty"`    // request body (for proxy)
	Command string            `json:"command,omitempty"` // shell command (for exec)
	WorkDir string            `json:"work_dir,omitempty"`
	Vault   string            `json:"vault,omitempty"` // named vault to use (empty for default)
}

// Response is the JSON response format from the daemon.
type Response struct {
	OK         bool              `json:"ok"`
	Error      string            `json:"error,omitempty"`
	Status     int               `json:"status,omitempty"`      // HTTP status (for proxy)
	Headers    map[string]string `json:"headers,omitempty"`     // response headers (for proxy)
	Body       string            `json:"body,omitempty"`        // response body (redacted)
	ExitCode   *int              `json:"exit_code,omitempty"`   // for exec
	Stdout     string            `json:"stdout,omitempty"`      // redacted stdout (for exec)
	Stderr     string            `json:"stderr,omitempty"`      // redacted stderr (for exec)
	SecretList []SecretInfo      `json:"secrets,omitempty"`     // for list
}

// SecretInfo is a non-sensitive description of a secret.
type SecretInfo struct {
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	AllowedDomains  []string `json:"allowed_domains,omitempty"`
	AllowedCommands []string `json:"allowed_commands,omitempty"`
	InjectAs        string   `json:"inject_as,omitempty"`
	Vault           string   `json:"vault,omitempty"`
}
