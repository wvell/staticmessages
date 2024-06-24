package staticmessages_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wvell/staticmessages"
)

var genGolden = flag.Bool("gen_golden", false, "Generate golden template files")

func TestWriteTemplateWithLocales(t *testing.T) {
	flag.Parse()

	defaultMsg, err := staticmessages.ParseMessage("Hello, %(user)s!")
	require.NoError(t, err)

	localizedOne, err := staticmessages.NewLocalizedMessage("HelloUser", defaultMsg)
	require.NoError(t, err)

	nlMsg, err := staticmessages.ParseMessage("Hallo, %(user)s, je hebt %(n)d! nieuwe berichten!")
	require.NoError(t, err)

	err = localizedOne.AddTranslation("nl", nlMsg)
	require.NoError(t, err)

	defaultMsg, err = staticmessages.ParseMessage("Hello world!")
	require.NoError(t, err)

	localizedTwo, err := staticmessages.NewLocalizedMessage("HelloWorld", defaultMsg)
	require.NoError(t, err)

	message := &staticmessages.Messages{
		Name: "Test",
		Messages: []*staticmessages.LocalizedMessage{
			localizedOne,
			localizedTwo,
		},
	}

	writeMessages(t, message, "template.golden_locales")
}

func TestWriteTemplateWithoutLocales(t *testing.T) {
	defaultMsg, err := staticmessages.ParseMessage("Hello %(user)s! Your cart has %(items)d and total is %(total).2f.")
	require.NoError(t, err)

	localized, err := staticmessages.NewLocalizedMessage("HelloWorld", defaultMsg)
	require.NoError(t, err)

	message := &staticmessages.Messages{
		Name: "Test",
		Messages: []*staticmessages.LocalizedMessage{
			localized,
		},
	}

	writeMessages(t, message, "template.golden_no_locales")
}

func writeMessages(t *testing.T, message *staticmessages.Messages, goldenFile string) {
	var buf bytes.Buffer
	err := staticmessages.Write(message, "testpkg", &buf)
	require.NoError(t, err)

	goldenPath := filepath.Join("./testdata/", goldenFile)
	if *genGolden {
		err = os.MkdirAll("./testdata", 0755)
		if err != nil {
			t.Fatalf("Failed to create the testdata directory: %v", err)
		}

		err = os.WriteFile(goldenPath, buf.Bytes(), 0644)
		if err != nil {
			t.Fatalf("Failed to write the golden file: %v", err)
		}

		return
	}

	golden, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	if !bytes.Equal(golden, buf.Bytes()) {
		t.Log("Golden:")
		t.Log(string(golden))

		t.Log("Generated:")
		t.Log(buf.String())
		t.Fatalf("Generated file does not match the golden file")
	}
}
