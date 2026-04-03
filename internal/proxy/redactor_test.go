package proxy

import (
	"encoding/base64"
	"net/url"
	"testing"
)

func TestRedactRawValue(t *testing.T) {
	r := NewRedactor(map[string]string{
		"API_KEY": "sk_live_abc123",
	})

	input := "Response contains sk_live_abc123 in the body"
	got := r.RedactString(input)
	want := "Response contains [VAULTY:API_KEY] in the body"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRedactBase64(t *testing.T) {
	secret := "sk_live_abc123"
	b64 := base64.StdEncoding.EncodeToString([]byte(secret))

	r := NewRedactor(map[string]string{
		"API_KEY": secret,
	})

	input := "Authorization: Basic " + b64
	got := r.RedactString(input)

	if got != "Authorization: Basic [VAULTY:API_KEY:b64]" {
		t.Errorf("got %q", got)
	}
}

func TestRedactURLEncoded(t *testing.T) {
	secret := "postgres://user:p@ss/db"
	urlEnc := url.QueryEscape(secret)

	r := NewRedactor(map[string]string{
		"DB_URL": secret,
	})

	input := "url=" + urlEnc
	got := r.RedactString(input)

	if got != "url=[VAULTY:DB_URL:url]" {
		t.Errorf("got %q", got)
	}
}

func TestRedactMultipleSecrets(t *testing.T) {
	r := NewRedactor(map[string]string{
		"KEY_A": "secret_a",
		"KEY_B": "secret_b",
	})

	input := "A=secret_a B=secret_b"
	got := r.RedactString(input)
	want := "A=[VAULTY:KEY_A] B=[VAULTY:KEY_B]"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRedactNoMatch(t *testing.T) {
	r := NewRedactor(map[string]string{
		"API_KEY": "sk_live_abc123",
	})

	input := "nothing to redact here"
	got := r.RedactString(input)

	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestRedactEmptySecret(t *testing.T) {
	r := NewRedactor(map[string]string{
		"EMPTY": "",
	})

	input := "should not crash"
	got := r.RedactString(input)

	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}
