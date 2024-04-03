package messages_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wvell/messages"
)

func TestParser(t *testing.T) {
	t.Run("name invalid", func(t *testing.T) {
		_, err := messages.Parse("", strings.NewReader("test"))
		require.ErrorIs(t, err, messages.ErrYamlNameInvalid)
	})

	t.Run("invalid yml", func(t *testing.T) {
		_, err := messages.Parse("Invalid", strings.NewReader(":"))
		require.ErrorIs(t, err, messages.ErrYamlDefinitionInvalid)
	})

	t.Run("invalid structure", func(t *testing.T) {
		_, err := messages.Parse("Invalid", strings.NewReader("Some structure"))
		require.ErrorIs(t, err, messages.ErrYamlDefinitionInvalid)
	})

	t.Run("identifier not capitalized", func(t *testing.T) {
		_, err := messages.Parse("Invalid", strings.NewReader(`someMessage:
  default: Hello, World!
someSecondMessage:
  default: Hello, World!
`))
		require.ErrorIs(t, err, messages.ErrYamlDefinitionInvalid)
	})

	t.Run("valid", func(t *testing.T) {
		container, err := messages.Parse("valid", strings.NewReader(`HelloWorld:
  default: Hello, World!
  nl: Hallo, Wereld!
HelloUser:
  default: Hello, %(user)s!
  nl: Hallo, %(user)s!
  de: Hallo!
`))
		require.NoError(t, err)

		require.Equal(t, "Valid", container.Name)
		require.Len(t, container.Messages, 2)

		require.Equal(t, "HelloWorld", container.Messages[0].Identifier)
		require.Equal(t, "Hello, World!", container.Messages[0].Default.Message)

		require.Equal(t, "HelloUser", container.Messages[1].Identifier)
		require.Equal(t, "user", container.Messages[1].Default.Vars[0].Name)
		require.Equal(t, "nl", container.Messages[1].Translations[0].Locale)
		require.Equal(t, "user", container.Messages[1].Translations[0].Message.Vars[0].Name)

		require.Equal(t, "de", container.Messages[1].Translations[1].Locale)
		require.Len(t, container.Messages[1].Translations[1].Message.Vars, 0)
	})
}
