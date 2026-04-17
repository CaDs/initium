package cron

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// entry holds a scheduled task.
type entry struct {
	interval time.Duration
	fn       func(ctx context.Context)
}

// Scheduler is a minimal in-process periodic scheduler. It uses a ticker per
// registered task — no external dependency required.
type Scheduler struct {
	entries []entry
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// New returns a ready-to-use Scheduler.
func New() *Scheduler {
	return &Scheduler{stopCh: make(chan struct{})}
}

// Every registers fn to be called every d. Must be called before Start.
func (s *Scheduler) Every(d time.Duration, fn func(ctx context.Context)) {
	s.entries = append(s.entries, entry{interval: d, fn: fn})
}

// Start launches all registered tasks in the background.
// The provided ctx is forwarded to each task invocation.
func (s *Scheduler) Start(ctx context.Context) {
	for _, e := range s.entries {
		e := e // capture
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ticker := time.NewTicker(e.interval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					func() {
						defer func() {
							if r := recover(); r != nil {
								slog.Error("cron task panicked", "panic", r)
							}
						}()
						e.fn(ctx)
					}()
				case <-s.stopCh:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// Stop signals all tasks to stop and waits for them to finish.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}
