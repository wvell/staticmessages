package messages_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wvell/messages"
)

func TestMessages(t *testing.T) {
	t.Run("duplicate identifiers", func(t *testing.T) {
		container := &messages.Messages{}

		defaultMessage, err := messages.ParseMessage("Hello, World!")
		require.NoError(t, err)

		loc, err := messages.NewLocalizedMessage("Foo", defaultMessage)
		require.NoError(t, err)

		err = container.Add(loc)
		require.NoError(t, err)

		err = container.Add(loc)
		require.ErrorIs(t, err, messages.ErrDuplicateIdentifier)
	})
}

func TestLocalizedMessage(t *testing.T) {
	defaultMessage, err := messages.ParseMessage("Hello, World!")
	require.NoError(t, err)

	t.Run("empty name", func(t *testing.T) {
		_, err := messages.NewLocalizedMessage("", defaultMessage)
		require.Equal(t, messages.ErrIdentifierInvalid, err, "expected error for empty name")
	})

	t.Run("name not uppercase", func(t *testing.T) {
		_, err := messages.NewLocalizedMessage("foo", defaultMessage)
		require.Equal(t, messages.ErrIdentifierInvalid, err, "expected error for name not uppercase")
	})

	t.Run("valid name", func(t *testing.T) {
		c, err := messages.NewLocalizedMessage("Foo", defaultMessage)
		require.NoError(t, err, "expected no error for valid name")
		require.Equal(t, "Foo", c.Identifier, "expected name to be set")
		require.NotNil(t, c.Translations, "expected translations to be initialized")
	})

	t.Run("invalid vars between default and translation", func(t *testing.T) {
		msg, err := messages.ParseMessage("Hello, %(user)s!")
		require.NoError(t, err)

		c, err := messages.NewLocalizedMessage("Foo", msg)
		require.NoError(t, err, "expected no error for valid name")

		tr, err := messages.ParseMessage("Hello, %(user)d!")
		require.NoError(t, err)

		err = c.AddTranslation("en", tr)
		require.ErrorIs(t, err, messages.ErrVariableTypeMix)
	})

	t.Run("invalid vars between 2 translations", func(t *testing.T) {
		msg, err := messages.ParseMessage("Hello, world!")
		require.NoError(t, err)

		c, err := messages.NewLocalizedMessage("Foo", msg)
		require.NoError(t, err, "expected no error for valid name")

		tr, err := messages.ParseMessage("Hallo, %(user)d!")
		require.NoError(t, err)

		err = c.AddTranslation("nl", tr)
		require.NoError(t, err)

		tr, err = messages.ParseMessage("Hallo, %(user)s!")
		require.NoError(t, err)

		err = c.AddTranslation("de", tr)
		require.ErrorIs(t, err, messages.ErrVariableTypeMix)
	})

	t.Run("duplicate translations", func(t *testing.T) {
		msg, err := messages.ParseMessage("Hello, %(user)s!")
		require.NoError(t, err)

		c, err := messages.NewLocalizedMessage("Foo", msg)
		require.NoError(t, err, "expected no error for valid name")

		tr, err := messages.ParseMessage("Hallo, %(user)s!")
		require.NoError(t, err)

		err = c.AddTranslation("nl", tr)
		require.NoError(t, err)

		err = c.AddTranslation("nl", tr)
		require.ErrorIs(t, err, messages.ErrDuplicateTranslation)
	})

	t.Run("valid translations", func(t *testing.T) {
		msg, err := messages.ParseMessage("Hello, %(user)s!")
		require.NoError(t, err)

		c, err := messages.NewLocalizedMessage("Foo", msg)
		require.NoError(t, err, "expected no error for valid name")

		tr, err := messages.ParseMessage("Hallo, %(user)s!")
		require.NoError(t, err)

		err = c.AddTranslation("nl", tr)
		require.NoError(t, err)
	})

	t.Run("unique translations", func(t *testing.T) {
		msg, err := messages.ParseMessage("Hello, %(user)s!")
		require.NoError(t, err)

		c, err := messages.NewLocalizedMessage("Foo", msg)
		require.NoError(t, err, "expected no error for valid name")

		tr, err := messages.ParseMessage("Hallo, %(user)s! Er zijn %(count)d nieuwe berichten!")
		require.NoError(t, err)

		err = c.AddTranslation("nl", tr)
		require.NoError(t, err)

		tr, err = messages.ParseMessage("Dein letzter Anmeldeversuch war vor %(tage)d Tagen.")
		require.NoError(t, err)

		err = c.AddTranslation("de", tr)
		require.NoError(t, err)

		unique := c.UniqueVars()
		require.Len(t, unique, 3)
	})
}

func TestParseMessage(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		tr, err := messages.ParseMessage("Hello, World!")
		require.NoError(t, err)
		require.Equal(t, "Hello, World!", tr.Message, "expected message to be set")
		require.Len(t, tr.Vars, 0)
	})

	t.Run("reserved keyword var", func(t *testing.T) {
		for _, reserved := range []string{
			"ctx",
			"break", "default", "func", "interface", "select",
			"case", "defer", "go", "map", "struct",
			"chan", "else", "goto", "package", "switch",
			"const", "fallthrough", "if", "range", "type",
			"continue", "for", "import", "return", "var",
		} {
			_, err := messages.ParseMessage("Hello, World! %(" + reserved + ")s")
			require.ErrorIs(t, err, messages.ErrReservedKeyword)
		}
	})

	t.Run("unsupported format", func(t *testing.T) {
		_, err := messages.ParseMessage("Hello, World! %(test)q")
		require.ErrorIs(t, err, messages.ErrUnsupportedFormat)
	})

	t.Run("same var twice in different format", func(t *testing.T) {
		_, err := messages.ParseMessage("Hello, %(test)d! %(test)s")
		require.ErrorIs(t, err, messages.ErrVariableTypeMix)
	})

	t.Run("parse vars", func(t *testing.T) {
		msg, err := messages.ParseMessage("Hello %(user)s! You have %(count)d new messages in your <a href=\"/user/%(user)s\">inbox</a>.")
		require.NoError(t, err)

		require.Equal(t, "Hello %s! You have %d new messages in your <a href=\"/user/%s\">inbox</a>.", msg.Message)
		require.Len(t, msg.Vars, 3)
		require.Len(t, msg.UniqueVars(), 2)
	})

	floatingPointCases := []struct {
		message     string
		expected    string
		expectedErr bool
	}{
		{
			message:  "Your order total is %(total)f!",
			expected: "Your order total is %f!",
		},
		{
			message:  "Your order total is %(total)9f!",
			expected: "Your order total is %9f!",
		},
		{
			message:  "Your order total is %(total)9.2f!",
			expected: "Your order total is %9.2f!",
		},
		{
			message:  "Your order total is %(total)9f!",
			expected: "Your order total is %9f!",
		},
		{
			message:  "Your order total is %(total).2f!",
			expected: "Your order total is %.2f!",
		},
		{
			message:     "Your order total is %(total)9..f!",
			expectedErr: true,
		},
	}

	for _, c := range floatingPointCases {
		t.Run(c.message, func(t *testing.T) {
			msg, err := messages.ParseMessage(c.message)
			if c.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expected, msg.Message)
			}
		})
	}
}
