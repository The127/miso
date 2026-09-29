//go:build vmtest

package place_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/place"
)

func TestAnAbsoluteLinkIsReadAsAPlaceInTheImage(t *testing.T) {
	// arrange
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "proc"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "proc", "version"), []byte("image\n"), 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(root, "boot"), 0o700))
	// the builder VM has a /proc/version of its own
	require.NoError(t, os.Symlink("/proc/version", filepath.Join(root, "boot", "vmlinuz")))

	// act
	read, err := fs.ReadFile(place.Open(root).FS(), "boot/vmlinuz")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "image\n", string(read))
}
