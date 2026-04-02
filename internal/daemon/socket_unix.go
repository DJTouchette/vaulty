//go:build !windows

package daemon

import (
	"log"
	"net"
	"net/http"
	"os"
)

// StartSocketListener creates a Unix domain socket listener and starts
// serving HTTP on it. Returns the listener and server for cleanup.
func StartSocketListener(socketPath string, handler http.Handler) (net.Listener, *http.Server, error) {
	os.Remove(socketPath) // clean up stale socket
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, nil, err
	}
	os.Chmod(socketPath, 0600)

	server := &http.Server{Handler: handler}
	go func() {
		log.Printf("Listening on %s", socketPath)
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("Socket server error: %v", err)
		}
	}()

	return ln, server, nil
}
