package worker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eridia/initium/backend/internal/infra/worker"
)

func TestPool_ProcessesAllJobs(t *testing.T) {
	t.Parallel()

	const jobCount = 10
	pool := worker.New()

	var processed atomic.Int64
	done := make(chan struct{})

	go func() {
		for range jobCount {
			pool.Submit(func(_ context.Context) {
				processed.Add(1)
			})
		}
		pool.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("pool did not process all jobs within 2s")
	}

	assert.Equal(t, int64(jobCount), processed.Load(),
		"every submitted job must be processed before Close returns")
}
