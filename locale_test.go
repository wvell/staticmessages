package messages_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wvell/messages"
	message "github.com/wvell/messages"
)

func TestLocaleContext(t *testing.T) {
	ctx := context.Background()

	require.Equal(t, "", messages.GetLocale(ctx))

	ctx = message.WrapLocale(ctx, "en-US")
	require.Equal(t, "en-US", message.GetLocale(ctx))
}
