package infra

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

// ShutdownHook is called during graceful shutdown, before the HTTP server stops.
// Implementations should complete their cleanup within the provided context deadline.
type ShutdownHook func(ctx context.Context)

// ServeHTTP starts the HTTP server with graceful shutdown on SIGTERM/SIGINT.
// hooks are invoked in order before the HTTP server shuts down; each shares a
// 5-second deadline to drain in-flight work.
func ServeHTTP(handler http.Handler, port int, hooks ...ShutdownHook) error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "port", port)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		slog.Info("shutting down gracefully...")

		// Run pre-shutdown hooks (worker drain, cron stop, etc.) with a 5s budget.
		hookCtx, hookCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer hookCancel()
		for _, hook := range hooks {
			hook(hookCtx)
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
		slog.Info("server stopped")
	}

	return nil
}
