package vault

import (
	"encoding/json"
	"fmt"

	"filippo.io/age"
)

// Export serializes vault secrets as JSON and encrypts the result.
// If recipients are provided, encrypts for those recipients using EncryptMulti.
// Otherwise encrypts with the passphrase only.
func Export(srcVault *Vault, passphrase string, recipients []age.Recipient) ([]byte, error) {
	data := vaultData{Secrets: make(map[string]string, len(srcVault.secrets))}
	for name, value := range srcVault.secrets {
		data.Secrets[name] = string(value)
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshaling vault for export: %w", err)
	}

	ciphertext, err := EncryptMulti(passphrase, recipients, plaintext)
	if err != nil {
		return nil, fmt.Errorf("encrypting export: %w", err)
	}

	return ciphertext, nil
}

// Import decrypts an exported vault snapshot and returns a Vault.
// It tries identity file decryption first (if identityFile is non-empty),
// then falls back to passphrase decryption.
func Import(data []byte, passphrase string, identityFile string) (*Vault, error) {
	var plaintext []byte
	var err error

	if identityFile != "" {
		plaintext, err = DecryptWithIdentity(identityFile, data)
	} else {
		plaintext, err = Decrypt(passphrase, data)
	}
	if err != nil {
		return nil, fmt.Errorf("decrypting import: %w", err)
	}

	return parseVaultData(plaintext)
}

// MergeVaults copies secrets from src into dst. If overwrite is false,
// existing keys in dst are not replaced. Returns the number of secrets merged.
func MergeVaults(dst, src *Vault, overwrite bool) int {
	count := 0
	for _, name := range src.List() {
		if !overwrite && dst.Has(name) {
			continue
		}
		val, _ := src.Get(name)
		dst.Set(name, val)
		count++
	}
	return count
}
