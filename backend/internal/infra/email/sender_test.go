package email

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSMTPSender_SendMagicLink_ContextCanceled_ReturnsContextError(t *testing.T) {
	t.Parallel()

	sender, err := NewSMTPSender("localhost", 1025, "noreply@example.com", "https://app.example.com", "initium")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = sender.SendMagicLink(ctx, "user@example.com", "token")

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestSMTPSender_MagicLinkData_EscapesToken(t *testing.T) {
	t.Parallel()

	sender, err := NewSMTPSender("localhost", 1025, "noreply@example.com", "https://app.example.com", "initium")
	require.NoError(t, err)

	data := sender.magicLinkData("a+b c")

	assert.Equal(t, "https://app.example.com/api/auth/verify?token=a%2Bb+c", data["Link"])
	assert.Equal(t, "initium://auth/verify?token=a%2Bb+c", data["AppLink"])
}
