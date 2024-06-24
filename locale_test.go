package staticmessages_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	message "github.com/wvell/staticmessages"
)

func TestLocaleContext(t *testing.T) {
	ctx := context.Background()

	require.Equal(t, "", message.GetLocale(ctx))

	ctx = message.WrapLocale(ctx, "en-US")
	require.Equal(t, "en-US", message.GetLocale(ctx))
}
