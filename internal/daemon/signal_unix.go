//go:build !windows

package daemon

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// WaitForSignal blocks until an OS signal (SIGTERM/SIGINT), idle timeout,
// or context cancellation occurs. Returns nil on clean shutdown.
func WaitForSignal(ctx context.Context, idleTimer *time.Timer, errCh <-chan error) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		log.Printf("Received %s, shutting down...", sig)
	case err := <-errCh:
		return err
	case <-idleTimer.C:
		log.Println("Idle timeout reached, shutting down...")
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	}

	return nil
}
