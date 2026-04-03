package vault

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// Encrypt encrypts plaintext using a passphrase via age's scrypt recipient.
func Encrypt(passphrase string, plaintext []byte) ([]byte, error) {
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return nil, fmt.Errorf("creating scrypt recipient: %w", err)
	}
	// Use a lower work factor for faster UX (still secure for local-only vault)
	recipient.SetWorkFactor(15)

	var buf bytes.Buffer
	armorWriter := armor.NewWriter(&buf)

	w, err := age.Encrypt(armorWriter, recipient)
	if err != nil {
		return nil, fmt.Errorf("creating encrypted writer: %w", err)
	}

	if _, err := w.Write(plaintext); err != nil {
		return nil, fmt.Errorf("writing plaintext: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("closing encrypted writer: %w", err)
	}

	if err := armorWriter.Close(); err != nil {
		return nil, fmt.Errorf("closing armor writer: %w", err)
	}

	return buf.Bytes(), nil
}

// EncryptMulti encrypts plaintext for both a passphrase (scrypt) and multiple
// age X25519 recipients. If recipients is empty, it falls back to passphrase-
// only encryption. When recipients are provided, the vault is encrypted for
// the X25519 recipients (age does not allow mixing scrypt with other recipient
// types in a single Encrypt call).
func EncryptMulti(passphrase string, recipients []age.Recipient, plaintext []byte) ([]byte, error) {
	if len(recipients) == 0 {
		return Encrypt(passphrase, plaintext)
	}

	// Encrypt for X25519 recipients only. Age's scrypt recipient cannot be
	// mixed with other recipient types, so team-shared vaults use identity-
	// based encryption exclusively.
	var buf bytes.Buffer
	armorWriter := armor.NewWriter(&buf)

	w, err := age.Encrypt(armorWriter, recipients...)
	if err != nil {
		return nil, fmt.Errorf("creating multi-recipient encrypted writer: %w", err)
	}
	if _, err := w.Write(plaintext); err != nil {
		return nil, fmt.Errorf("writing plaintext: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("closing encrypted writer: %w", err)
	}
	if err := armorWriter.Close(); err != nil {
		return nil, fmt.Errorf("closing armor writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Decrypt decrypts age-encrypted ciphertext using a passphrase.
func Decrypt(passphrase string, ciphertext []byte) ([]byte, error) {
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, fmt.Errorf("creating scrypt identity: %w", err)
	}

	armorReader := armor.NewReader(bytes.NewReader(ciphertext))

	r, err := age.Decrypt(armorReader, identity)
	if err != nil {
		return nil, fmt.Errorf("decrypting vault (wrong passphrase?): %w", err)
	}

	plaintext, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading decrypted data: %w", err)
	}

	return plaintext, nil
}

// DecryptWithIdentity decrypts age-encrypted ciphertext using an age identity
// file (private key file).
func DecryptWithIdentity(identityFile string, ciphertext []byte) ([]byte, error) {
	ids, err := parseIdentitiesFromFile(identityFile)
	if err != nil {
		return nil, err
	}

	armorReader := armor.NewReader(bytes.NewReader(ciphertext))

	r, err := age.Decrypt(armorReader, ids...)
	if err != nil {
		return nil, fmt.Errorf("decrypting with identity file: %w", err)
	}

	plaintext, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading decrypted data: %w", err)
	}

	return plaintext, nil
}

// parseIdentitiesFromFile reads an age identity file and returns the identities.
func parseIdentitiesFromFile(path string) ([]age.Identity, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening identity file: %w", err)
	}
	defer f.Close()

	ids, err := age.ParseIdentities(f)
	if err != nil {
		return nil, fmt.Errorf("parsing identity file: %w", err)
	}

	return ids, nil
}
