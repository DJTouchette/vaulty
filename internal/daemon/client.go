package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// Client communicates with the running daemon.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewSocketClient creates a client that connects via Unix socket.
func NewSocketClient(socketPath string) *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return net.DialTimeout("unix", socketPath, 5*time.Second)
				},
			},
			Timeout: 60 * time.Second,
		},
		baseURL: "http://vaulty",
	}
}

// NewHTTPClient creates a client that connects via localhost HTTP.
func NewHTTPClient(port int) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
	}
}

// NewClient tries to connect via socket first, falls back to HTTP.
func NewClient(socketPath string, httpPort int) *Client {
	// Try socket first
	if socketPath != "" {
		conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
		if err == nil {
			conn.Close()
			return NewSocketClient(socketPath)
		}
	}

	// Fall back to HTTP
	return NewHTTPClient(httpPort)
}

// Send sends a request to the daemon and returns the response.
func (c *Client) Send(req Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/v1/request", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("daemon not running or unreachable: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &resp, nil
}
