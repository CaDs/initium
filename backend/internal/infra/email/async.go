package email

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/eridia/initium/backend/internal/domain"
	"github.com/eridia/initium/backend/internal/infra/worker"
)

// AsyncSender wraps an EmailSender and dispatches sends to a worker pool,
// returning nil immediately so the HTTP request is not blocked.
type AsyncSender struct {
	pool  *worker.Pool
	inner domain.EmailSender
}

// NewAsyncSender creates an AsyncSender that submits email jobs to pool.
func NewAsyncSender(pool *worker.Pool, inner domain.EmailSender) *AsyncSender {
	return &AsyncSender{pool: pool, inner: inner}
}

// SendMagicLink submits the email delivery to the worker pool and returns nil
// immediately. Delivery failures are logged but not propagated to the caller.
func (a *AsyncSender) SendMagicLink(ctx context.Context, to string, token string) error {
	// Capture values; do not close over ctx (it may be cancelled by request end).
	submitted := a.pool.Submit(func(jobCtx context.Context) {
		if err := a.inner.SendMagicLink(jobCtx, to, token); err != nil {
			slog.Error("async email delivery failed", "to", to, "error", err)
		}
	})
	if !submitted {
		slog.Warn("magic link email dropped: worker queue full", "to", to)
		return fmt.Errorf("queueing magic link email: worker queue full")
	}
	return nil
}
