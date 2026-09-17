package buildcontext_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/buildcontext"
)

func open(t *testing.T, dir string) *buildcontext.Dir {
	t.Helper()

	context, err := buildcontext.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, context.Close()) })

	return context
}

func write(t *testing.T, dir string, name string, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}

func TestABuildContextThatIsNotThereCannotBeOpened(t *testing.T) {
	// arrange
	missing := filepath.Join(t.TempDir(), "nope")

	// act
	_, err := buildcontext.Open(missing)

	// assert
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestAMissingPathDoesNotExist(t *testing.T) {
	// arrange
	context := open(t, t.TempDir())

	// act
	_, err := context.Digest("nope")

	// assert
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestAChangedFileContentChangesTheDigest(t *testing.T) {
	// arrange
	hello := t.TempDir()
	write(t, hello, "motd", "hello")
	goodbye := t.TempDir()
	write(t, goodbye, "motd", "goodbye")

	// act
	helloDigest, helloErr := open(t, hello).Digest("motd")
	goodbyeDigest, goodbyeErr := open(t, goodbye).Digest("motd")

	// assert
	require.NoError(t, helloErr)
	require.NoError(t, goodbyeErr)
	assert.NotEqual(t, helloDigest, goodbyeDigest)
}
