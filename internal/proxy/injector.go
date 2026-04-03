package proxy

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
)

// InjectMode defines how a secret is injected into a request.
type InjectMode string

const (
	InjectBearer InjectMode = "bearer"
	InjectBasic  InjectMode = "basic"
	InjectHeader InjectMode = "header"
	InjectQuery  InjectMode = "query"
)

// InjectSecret modifies the HTTP request to include the secret based on the injection mode.
func InjectSecret(req *http.Request, secret string, mode InjectMode, headerName string) error {
	switch mode {
	case InjectBearer, "":
		req.Header.Set("Authorization", "Bearer "+secret)
	case InjectBasic:
		encoded := base64.StdEncoding.EncodeToString([]byte(secret))
		req.Header.Set("Authorization", "Basic "+encoded)
	case InjectHeader:
		if headerName == "" {
			return fmt.Errorf("header_name required for inject_as=header")
		}
		req.Header.Set(headerName, secret)
	case InjectQuery:
		q := req.URL.Query()
		q.Set("key", secret)
		req.URL.RawQuery = q.Encode()
	default:
		return fmt.Errorf("unknown inject_as mode: %q", mode)
	}
	return nil
}

// InjectSecretIntoURL is used for the query injection mode (for display/testing).
func InjectSecretIntoURL(rawURL, secret string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("key", secret)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
