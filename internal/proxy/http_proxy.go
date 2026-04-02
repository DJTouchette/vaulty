package proxy

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ProxyResult holds the result of a proxied HTTP request.
type ProxyResult struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

// DoRequest makes an authenticated HTTP request, injecting the secret and redacting the response.
func DoRequest(method, url string, headers map[string]string, body string, secret string, mode InjectMode, headerName string, redactor *Redactor) (*ProxyResult, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set user-provided headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Inject the secret
	if err := InjectSecret(req, secret, mode, headerName); err != nil {
		return nil, fmt.Errorf("injecting secret: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Redact any secret values from the response body
	redactedBody := string(redactor.Redact(respBody))

	// Collect response headers
	respHeaders := make(map[string]string)
	for k := range resp.Header {
		respHeaders[k] = redactor.RedactString(resp.Header.Get(k))
	}

	return &ProxyResult{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       redactedBody,
	}, nil
}
