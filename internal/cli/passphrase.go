package cli

import (
	"fmt"
	"os"

	"github.com/djtouchette/vaulty/internal/vault"
	"golang.org/x/term"
)

// vaultHandle holds an opened vault and the passphrase used (empty if identity-based).
type vaultHandle struct {
	Vault      *vault.Vault
	Passphrase string
}

// resolveVaultPath returns the vault file path, taking the --vault flag into account.
func resolveVaultPath(basePath string) string {
	return vault.ResolveVaultPath(vaultName, basePath)
}

// openVault opens the vault using the best available method:
// 1. --identity flag (age private key file)
// 2. VAULTY_IDENTITY env var
// 3. OS keychain passphrase
// 4. VAULTY_PASSPHRASE env var
// 5. Interactive passphrase prompt
func openVault(vaultPath string) (*vaultHandle, error) {
	// Try identity file first (team mode)
	identity := identityFile
	if identity == "" {
		identity = os.Getenv("VAULTY_IDENTITY")
	}
	if identity != "" {
		v, err := vault.OpenWithIdentity(vaultPath, identity)
		if err != nil {
			return nil, err
		}
		return &vaultHandle{Vault: v, Passphrase: ""}, nil
	}

	// Fall back to passphrase
	pass, err := getPassphrase(vaultPath)
	if err != nil {
		return nil, err
	}
	v, err := vault.Open(vaultPath, pass)
	if err != nil {
		return nil, err
	}
	return &vaultHandle{Vault: v, Passphrase: pass}, nil
}

// getPassphrase tries to retrieve the vault passphrase from (in order):
// 1. OS keychain
// 2. VAULTY_PASSPHRASE env var
// 3. Interactive terminal prompt
func getPassphrase(vaultPath string) (string, error) {
	service := vault.DefaultService()
	account := vault.KeyringAccount(vaultPath)

	if pass, err := vault.GetPassphrase(service, account); err == nil {
		return pass, nil
	}

	if pass := os.Getenv("VAULTY_PASSPHRASE"); pass != "" {
		return pass, nil
	}

	// Fall back to interactive prompt
	fmt.Print("Passphrase: ")
	raw, err := term.ReadPassword(0)
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	return string(raw), nil
}
