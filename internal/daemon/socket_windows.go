//go:build windows

package daemon

import (
	"log"
	"net"
	"net/http"
)

// StartSocketListener is a no-op on Windows. Unix domain sockets are not
// available, so the daemon operates in HTTP-only mode.
func StartSocketListener(socketPath string, handler http.Handler) (net.Listener, *http.Server, error) {
	log.Println("Unix sockets are not available on Windows; using HTTP-only mode")
	return nil, nil, nil
}
