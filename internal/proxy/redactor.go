package proxy

import (
	"bytes"
	"encoding/base64"
	"net/url"
)

// Redactor replaces secret values in output with safe placeholders.
type Redactor struct {
	// replacements maps the raw bytes to find -> replacement label
	replacements []replacement
}

type replacement struct {
	find    []byte
	replace []byte
}

// NewRedactor creates a redactor for the given secret name/value pairs.
func NewRedactor(secrets map[string]string) *Redactor {
	var reps []replacement

	for name, value := range secrets {
		if value == "" {
			continue
		}

		// Raw value
		reps = append(reps, replacement{
			find:    []byte(value),
			replace: []byte("[VAULTY:" + name + "]"),
		})

		// Base64-encoded
		b64 := base64.StdEncoding.EncodeToString([]byte(value))
		if b64 != value {
			reps = append(reps, replacement{
				find:    []byte(b64),
				replace: []byte("[VAULTY:" + name + ":b64]"),
			})
		}

		// URL-encoded
		urlEnc := url.QueryEscape(value)
		if urlEnc != value {
			reps = append(reps, replacement{
				find:    []byte(urlEnc),
				replace: []byte("[VAULTY:" + name + ":url]"),
			})
		}
	}

	// Sort by length descending so longer matches are replaced first
	// (prevents partial matches from masking longer ones)
	for i := 0; i < len(reps); i++ {
		for j := i + 1; j < len(reps); j++ {
			if len(reps[j].find) > len(reps[i].find) {
				reps[i], reps[j] = reps[j], reps[i]
			}
		}
	}

	return &Redactor{replacements: reps}
}

// Redact replaces all known secret values in the input.
func (r *Redactor) Redact(input []byte) []byte {
	result := input
	for _, rep := range r.replacements {
		result = bytes.ReplaceAll(result, rep.find, rep.replace)
	}
	return result
}

// RedactString is a convenience wrapper for string input.
func (r *Redactor) RedactString(input string) string {
	return string(r.Redact([]byte(input)))
}
