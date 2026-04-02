package backend

// SecretBackend is the interface for external secret providers.
type SecretBackend interface {
	// Name returns the backend identifier (e.g., "aws-secrets-manager").
	Name() string
	// List returns available secret names.
	List() ([]string, error)
	// Get retrieves a secret value by name.
	Get(name string) (string, error)
}
