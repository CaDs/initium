package worker

import (
	"context"
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
}

// New creates a Pool with defaultWorkers workers and a job queue of defaultQueue.
// Call Close to drain all queued jobs and wait for workers to finish.
func New() *Pool {
	ctx, cancel := context.WithCancel(context.Background())
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
	select {
	case p.jobs <- job:
		return true
	default:
		slog.Warn("worker pool queue full, dropping job")
		return false
	}
}

// Close signals no more jobs, waits for in-flight work to finish, then returns.
// A context with a deadline should be used by the caller to bound the wait.
func (p *Pool) Close() {
	close(p.jobs)
	p.wg.Wait()
	p.cancel()
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
