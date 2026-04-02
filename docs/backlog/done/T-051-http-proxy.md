# T-051: HTTP proxy

**Epic:** 5 — HTTP Proxy & Secret Injection
**Status:** done
**Priority:** P0

## Description

Implement `internal/proxy/http_proxy.go` — make authenticated HTTP requests.

## Acceptance Criteria

- [ ] Accepts method, URL, headers, body
- [ ] Injects secret via injector
- [ ] Makes HTTP request and returns response (status, headers, body)
- [ ] Redacts secret values from response body before returning
- [ ] Respects reasonable timeouts
- [ ] Integration test against httptest server
