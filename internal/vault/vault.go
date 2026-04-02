package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type vaultData struct {
	Secrets map[string]string `json:"secrets"`
}

// Vault holds decrypted secrets in memory.
type Vault struct {
	secrets map[string][]byte
}

// ResolveVaultPath returns the file path for a named vault. If name is empty,
// returns basePath unchanged (the default vault). If name is provided, returns
// <dir>/vaults/<name>.age where <dir> is the parent directory of basePath.
func ResolveVaultPath(name, basePath string) string {
	if name == "" {
		return basePath
	}
	dir := filepath.Dir(expandPath(basePath))
	return filepath.Join(dir, "vaults", name+".age")
}

// Exists returns true if a vault file exists at the given path.
func Exists(path string) bool {
	path = expandPath(path)
	_, err := os.Stat(path)
	return err == nil
}

// Create creates a new empty vault file encrypted with the given passphrase.
func Create(path, passphrase string) error {
	path = expandPath(path)

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating vault directory: %w", err)
	}

	data := vaultData{Secrets: map[string]string{}}
	plaintext, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling vault: %w", err)
	}

	ciphertext, err := Encrypt(passphrase, plaintext)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, ciphertext, 0600); err != nil {
		return fmt.Errorf("writing vault file: %w", err)
	}

	return nil
}

// Open decrypts the vault file using a passphrase and returns a Vault with secrets in memory.
func Open(path, passphrase string) (*Vault, error) {
	path = expandPath(path)

	ciphertext, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading vault file: %w", err)
	}

	plaintext, err := Decrypt(passphrase, ciphertext)
	if err != nil {
		return nil, err
	}

	return parseVaultData(plaintext)
}

// OpenWithIdentity decrypts the vault file using an age identity (private key) file.
func OpenWithIdentity(path, identityFile string) (*Vault, error) {
	path = expandPath(path)

	ciphertext, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading vault file: %w", err)
	}

	plaintext, err := DecryptWithIdentity(identityFile, ciphertext)
	if err != nil {
		return nil, err
	}

	return parseVaultData(plaintext)
}

func parseVaultData(plaintext []byte) (*Vault, error) {
	var data vaultData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("parsing vault data: %w", err)
	}

	// Store as []byte for secure zeroing
	v := &Vault{secrets: make(map[string][]byte, len(data.Secrets))}
	for name, value := range data.Secrets {
		v.secrets[name] = []byte(value)
	}

	return v, nil
}

// Save encrypts the vault data and writes it to disk.
// If team recipients exist for this vault, it encrypts for all recipients
// using age X25519 keys. Otherwise it uses passphrase-based encryption.
func (v *Vault) Save(path, passphrase string) error {
	path = expandPath(path)

	data := vaultData{Secrets: make(map[string]string, len(v.secrets))}
	for name, value := range v.secrets {
		data.Secrets[name] = string(value)
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling vault: %w", err)
	}

	// Check for team recipients
	recipients, err := LoadRecipients(path)
	if err != nil {
		return fmt.Errorf("loading recipients: %w", err)
	}

	ciphertext, err := EncryptMulti(passphrase, recipients, plaintext)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, ciphertext, 0600); err != nil {
		return fmt.Errorf("writing vault file: %w", err)
	}

	return nil
}

// Set adds or updates a secret.
func (v *Vault) Set(name, value string) {
	v.secrets[name] = []byte(value)
}

// Get retrieves a secret value. Returns ("", false) if not found.
func (v *Vault) Get(name string) (string, bool) {
	val, ok := v.secrets[name]
	if !ok {
		return "", false
	}
	return string(val), true
}

// Has returns true if the secret exists.
func (v *Vault) Has(name string) bool {
	_, ok := v.secrets[name]
	return ok
}

// Remove deletes a secret.
func (v *Vault) Remove(name string) {
	if val, ok := v.secrets[name]; ok {
		zeroBytes(val)
		delete(v.secrets, name)
	}
}

// List returns all secret names.
func (v *Vault) List() []string {
	names := make([]string, 0, len(v.secrets))
	for name := range v.secrets {
		names = append(names, name)
	}
	return names
}

// Zero securely zeroes all secret values in memory.
func (v *Vault) Zero() {
	for name, val := range v.secrets {
		zeroBytes(val)
		delete(v.secrets, name)
	}
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func expandPath(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
