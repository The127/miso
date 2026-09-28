package contextfiles_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/buildcontext"
	"github.com/The127/miso/internal/contextfiles"
	"github.com/The127/miso/internal/protocol"
)

// opened is a build context in a fresh directory with these files in it.
func opened(t *testing.T, files map[string]string) *buildcontext.Dir {
	t.Helper()

	dir := t.TempDir()
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}

	context, err := buildcontext.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = context.Close() })

	return context
}

// sent is a sent entry with its content.
type sent struct {
	entry   protocol.Entry
	content string
}

// sending collects what files send.
func sending(t *testing.T, files protocol.Files) ([]sent, error) {
	t.Helper()

	var got []sent
	err := files(func(entry protocol.Entry, content io.Reader) error {
		read, err := io.ReadAll(content)
		if err != nil {
			return err
		}

		got = append(got, sent{entry, string(read)})

		return nil
	})

	return got, err
}

func TestAFileIsSentAsTheFirstSourceWithItsContent(t *testing.T) {
	// arrange
	context := opened(t, map[string]string{"motd": "hello\n"})
	digest, err := context.Digest("motd")
	require.NoError(t, err)
	request := protocol.Copy{Key: "step", Sources: []string{"motd"}, Digests: []string{digest}, Destination: "/etc/motd"}

	// act
	got, err := sending(t, contextfiles.Of(context)(request))

	// assert
	require.NoError(t, err)
	assert.Equal(t, []sent{{protocol.Entry{Source: 0, Kind: "file", Path: ".", Mode: 0o600, Size: 6}, "hello\n"}}, got)
}

func TestEverySourceIsSentInOrderUnderItsIndex(t *testing.T) {
	// arrange
	context := opened(t, map[string]string{"motd": "hello\n", "issue": "hey\n"})
	motd, err := context.Digest("motd")
	require.NoError(t, err)
	issue, err := context.Digest("issue")
	require.NoError(t, err)
	request := protocol.Copy{Key: "step", Sources: []string{"motd", "issue"}, Digests: []string{motd, issue}, Destination: "/etc/"}

	// act
	got, err := sending(t, contextfiles.Of(context)(request))

	// assert
	require.NoError(t, err)
	assert.Equal(t, []sent{
		{protocol.Entry{Source: 0, Kind: "file", Path: ".", Mode: 0o600, Size: 6}, "hello\n"},
		{protocol.Entry{Source: 1, Kind: "file", Path: ".", Mode: 0o600, Size: 4}, "hey\n"},
	}, got)
}
