package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

const (
	defaultWorkers = 4
	defaultQueue   = 100
)

// Job is a unit of work submitted to the pool.
type Job func(ctx context.Context)

// Pool is a fixed-size goroutine pool with a bounded job queue.
type Pool struct {
	jobs   chan Job
	wg     sync.WaitGroup
	cancel context.CancelFunc
	mu     sync.RWMutex
	closed bool
}

// New creates a Pool with defaultWorkers workers and a job queue of defaultQueue.
// Call Close to drain all queued jobs and wait for workers to finish.
func New() *Pool {
	ctx, cancel := context.WithCancel(context.Background()) //nolint:gosec // Cancel is retained and called by Close.
	p := &Pool{
		jobs:   make(chan Job, defaultQueue),
		cancel: cancel,
	}
	for range defaultWorkers {
		p.wg.Add(1)
		go p.work(ctx)
	}
	return p
}

// Submit enqueues a job. Returns false if the queue is full and the job was dropped.
func (p *Pool) Submit(job Job) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		slog.Warn("worker pool closed, dropping job")
		return false
	}
	select {
	case p.jobs <- job:
		return true
	default:
		slog.Warn("worker pool queue full, dropping job")
		return false
	}
}

// Close signals no more jobs, waits for in-flight work to finish, then returns.
// The provided context bounds the drain wait.
func (p *Pool) Close(ctx context.Context) error {
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.jobs)
	}
	p.mu.Unlock()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.cancel()
		return nil
	case <-ctx.Done():
		p.cancel()
		return fmt.Errorf("closing worker pool: %w", ctx.Err())
	}
}

func (p *Pool) work(ctx context.Context) {
	defer p.wg.Done()
	for job := range p.jobs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("worker job panicked", "panic", r)
				}
			}()
			job(ctx)
		}()
	}
}
