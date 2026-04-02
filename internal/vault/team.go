package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
)

const recipientsFileName = "recipients"

// recipientsDir returns the .vaulty directory next to the vault file.
func recipientsDir(vaultPath string) string {
	vaultPath = expandPath(vaultPath)
	return filepath.Join(filepath.Dir(vaultPath), ".vaulty")
}

// recipientsFile returns the path to the recipients file.
func recipientsFile(vaultPath string) string {
	return filepath.Join(recipientsDir(vaultPath), recipientsFileName)
}

// AddRecipient adds an age X25519 public key as a recipient for the vault.
// The identityFileOrKey parameter can be either a raw public key string
// (starting with "age1") or a path to a file containing a public key.
func AddRecipient(vaultPath string, identityFileOrKey string) error {
	pubkey, err := resolvePublicKey(identityFileOrKey)
	if err != nil {
		return fmt.Errorf("resolving public key: %w", err)
	}

	// Validate the public key by parsing it
	if _, err := age.ParseX25519Recipient(pubkey); err != nil {
		return fmt.Errorf("invalid age public key %q: %w", pubkey, err)
	}

	// Check for duplicates
	existing, err := ListRecipients(vaultPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading existing recipients: %w", err)
	}
	for _, r := range existing {
		if r == pubkey {
			return fmt.Errorf("recipient %s already exists", pubkey)
		}
	}

	// Ensure the .vaulty directory exists
	dir := recipientsDir(vaultPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating recipients directory: %w", err)
	}

	// Append the public key to the recipients file
	f, err := os.OpenFile(recipientsFile(vaultPath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening recipients file: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, pubkey); err != nil {
		return fmt.Errorf("writing recipient: %w", err)
	}

	return nil
}

// ListRecipients returns all age public keys from the recipients file.
func ListRecipients(vaultPath string) ([]string, error) {
	path := recipientsFile(vaultPath)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening recipients file: %w", err)
	}
	defer f.Close()

	var recipients []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		recipients = append(recipients, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading recipients file: %w", err)
	}

	return recipients, nil
}

// RemoveRecipient removes a public key from the recipients file.
func RemoveRecipient(vaultPath string, pubkey string) error {
	existing, err := ListRecipients(vaultPath)
	if err != nil {
		return fmt.Errorf("reading recipients: %w", err)
	}

	found := false
	var remaining []string
	for _, r := range existing {
		if r == pubkey {
			found = true
			continue
		}
		remaining = append(remaining, r)
	}

	if !found {
		return fmt.Errorf("recipient %s not found", pubkey)
	}

	path := recipientsFile(vaultPath)
	if len(remaining) == 0 {
		return os.Remove(path)
	}

	content := strings.Join(remaining, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing recipients file: %w", err)
	}

	return nil
}

// LoadRecipients parses the recipients file and returns age.Recipient values.
func LoadRecipients(vaultPath string) ([]age.Recipient, error) {
	keys, err := ListRecipients(vaultPath)
	if err != nil {
		return nil, err
	}

	var recipients []age.Recipient
	for _, key := range keys {
		r, err := age.ParseX25519Recipient(key)
		if err != nil {
			return nil, fmt.Errorf("parsing recipient %q: %w", key, err)
		}
		recipients = append(recipients, r)
	}

	return recipients, nil
}

// resolvePublicKey takes either a raw age public key or a file path and
// returns the public key string.
func resolvePublicKey(input string) (string, error) {
	// If it looks like a raw public key, return it directly
	if strings.HasPrefix(input, "age1") {
		return input, nil
	}

	// Otherwise, try to read it as a file
	path := expandPath(input)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading key file %s: %w", path, err)
	}

	// Scan for a public key line in the file
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "age1") {
			return line, nil
		}
		// Check for a "# public key:" comment line (standard age key format)
		if strings.HasPrefix(line, "# public key: age1") {
			return strings.TrimPrefix(line, "# public key: "), nil
		}
	}

	return "", fmt.Errorf("no age public key found in %s", path)
}
