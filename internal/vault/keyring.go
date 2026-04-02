package vault

import (
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

const keyringService = "vaulty"

// KeyringAccount returns a keyring account name derived from the vault path.
// This allows different vaults to have separate keychain entries.
func KeyringAccount(vaultPath string) string {
	vaultPath = expandPath(vaultPath)
	abs, err := filepath.Abs(vaultPath)
	if err != nil {
		abs = vaultPath
	}
	return "vault:" + abs
}

// SavePassphrase stores the vault passphrase in the OS keychain.
func SavePassphrase(service, account, passphrase string) error {
	return keyring.Set(service, account, passphrase)
}

// GetPassphrase retrieves the vault passphrase from the OS keychain.
// Returns the passphrase and nil on success, or ("", error) if not found
// or the keychain is unavailable.
func GetPassphrase(service, account string) (string, error) {
	return keyring.Get(service, account)
}

// DeletePassphrase removes the vault passphrase from the OS keychain.
func DeletePassphrase(service, account string) error {
	return keyring.Delete(service, account)
}

// HasPassphrase checks whether a passphrase is stored in the OS keychain.
func HasPassphrase(service, account string) bool {
	_, err := keyring.Get(service, account)
	return err == nil
}

// IsKeyringAvailable returns true if the OS keychain backend is functional.
// It performs a harmless probe to detect missing backends (e.g. headless CI).
func IsKeyringAvailable() bool {
	// Try to get a key that almost certainly doesn't exist.
	// ErrNotFound means the backend works; any other error means it doesn't.
	_, err := keyring.Get(keyringService, "__vaulty_probe__")
	if err == nil {
		return true
	}
	return strings.Contains(err.Error(), "not found") || err == keyring.ErrNotFound
}

// DefaultService returns the default keyring service name.
func DefaultService() string {
	return keyringService
}
