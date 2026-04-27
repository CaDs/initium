package email_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/infra/email"
	"github.com/eridia/initium/backend/internal/infra/worker"
)

type stubSender struct{}

func (stubSender) SendMagicLink(context.Context, string, string) error {
	return nil
}

func TestAsyncSender_SendMagicLink_QueueFull_ReturnsError(t *testing.T) {
	t.Parallel()

	pool := worker.New()
	release := make(chan struct{})
	started := make(chan struct{}, 4)

	for range 4 {
		require.True(t, pool.Submit(func(context.Context) {
			started <- struct{}{}
			<-release
		}))
	}
	for range 4 {
		<-started
	}

	for pool.Submit(func(context.Context) { <-release }) {
	}

	sender := email.NewAsyncSender(pool, stubSender{})
	err := sender.SendMagicLink(context.Background(), "user@example.com", "token")

	require.Error(t, err)

	close(release)
	require.NoError(t, pool.Close(context.Background()))
}
