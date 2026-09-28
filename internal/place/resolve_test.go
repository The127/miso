package place_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/place"
)

func TestAnAbsoluteLinkInTheImageStaysInTheImage(t *testing.T) {
	// arrange
	root := t.TempDir()
	outside := t.TempDir()
	// the link means outside as the image sees it, below its own root
	require.NoError(t, os.MkdirAll(filepath.Join(root, outside), 0o700))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "etc")))

	// act
	err := place.Open(root).File("/etc/motd", 0o644, strings.NewReader("hello\n"))

	// assert
	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(outside, "motd"))
	assert.FileExists(t, filepath.Join(root, outside, "motd"))
}
