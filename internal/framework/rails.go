package framework

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseRailsCredentials parses YAML data (e.g., from decrypted Rails credentials)
// and flattens nested keys to SECTION_KEY format (uppercase, underscore separated).
// For example:
//
//	aws:
//	  access_key_id: AKID
//
// becomes AWS_ACCESS_KEY_ID=AKID.
func ParseRailsCredentials(data []byte) (map[string]string, error) {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing YAML credentials: %w", err)
	}
	return FlattenYAML("", raw), nil
}

// FlattenYAML recursively flattens a nested map into uppercase underscore-separated keys.
func FlattenYAML(prefix string, data map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for key, val := range data {
		fullKey := strings.ToUpper(key)
		if prefix != "" {
			fullKey = prefix + "_" + fullKey
		}

		switch v := val.(type) {
		case map[string]interface{}:
			for k, v := range FlattenYAML(fullKey, v) {
				result[k] = v
			}
		case map[interface{}]interface{}:
			// yaml.v3 shouldn't produce this, but handle defensively
			converted := make(map[string]interface{}, len(v))
			for mk, mv := range v {
				converted[fmt.Sprintf("%v", mk)] = mv
			}
			for k, v := range FlattenYAML(fullKey, converted) {
				result[k] = v
			}
		default:
			result[fullKey] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

// WriteRailsCredentials takes flattened secrets (SECTION_KEY format) and produces
// nested YAML output suitable for Rails credentials. For example:
//
//	AWS_ACCESS_KEY_ID=AKID -> aws:\n  access_key_id: AKID
func WriteRailsCredentials(secrets map[string]string) ([]byte, error) {
	nested := unflattenToNested(secrets)

	data, err := yaml.Marshal(nested)
	if err != nil {
		return nil, fmt.Errorf("marshaling Rails credentials YAML: %w", err)
	}
	return data, nil
}

// unflattenToNested converts uppercase underscore-separated keys back to a nested map
// with lowercase keys. Keys are split on underscore boundaries to create nesting.
// Single-segment keys remain at the top level.
func unflattenToNested(secrets map[string]string) map[string]interface{} {
	root := make(map[string]interface{})

	// Sort keys for deterministic output
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		parts := strings.Split(strings.ToLower(key), "_")
		val := secrets[key]

		if len(parts) == 1 {
			root[parts[0]] = val
			continue
		}

		// Use first segment as section, rest joined as the leaf key
		section := parts[0]
		leafKey := strings.Join(parts[1:], "_")

		sub, ok := root[section].(map[string]interface{})
		if !ok {
			sub = make(map[string]interface{})
			root[section] = sub
		}
		sub[leafKey] = val
	}

	return root
}

// DecryptRailsCredentials decrypts a Rails credentials file.
// Rails 7+ format: base64(encrypted_data)--base64(iv)--base64(auth_tag)
// The master key is a 32-byte hex string (64 hex chars) for AES-256-GCM.
func DecryptRailsCredentials(encPath, keyPath string) ([]byte, error) {
	encData, err := os.ReadFile(encPath)
	if err != nil {
		return nil, fmt.Errorf("reading encrypted credentials %s: %w", encPath, err)
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading master key %s: %w", keyPath, err)
	}

	masterKeyHex := strings.TrimSpace(string(keyData))
	masterKey, err := hex.DecodeString(masterKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decoding master key hex: %w (is %s a valid hex string?)", err, keyPath)
	}

	if len(masterKey) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes (got %d) — Rails 7+ uses AES-256-GCM", len(masterKey))
	}

	return decryptRailsPayload(strings.TrimSpace(string(encData)), masterKey)
}

func decryptRailsPayload(payload string, key []byte) ([]byte, error) {
	parts := strings.Split(payload, "--")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid Rails credentials format: expected 3 parts separated by '--', got %d", len(parts))
	}

	encryptedData, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decoding encrypted data: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding IV: %w", err)
	}

	authTag, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decoding auth tag: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	// GCM expects ciphertext with auth tag appended
	ciphertext := append(encryptedData, authTag...)

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting Rails credentials: %w (wrong master key?)", err)
	}

	return plaintext, nil
}
