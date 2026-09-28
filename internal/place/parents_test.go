//go:build vmtest

package place_test

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/place"
)

func TestMissingParentsOfAFileAreMadeOpenToAllAndRoots(t *testing.T) {
	// arrange
	root := t.TempDir()
	// a umask that would keep 0755 from any directory made without care
	old := syscall.Umask(0o077)
	t.Cleanup(func() { syscall.Umask(old) })

	// act
	err := place.Open(root).File("/usr/local/share/motd", 0o644, strings.NewReader("hello\n"))

	// assert
	require.NoError(t, err)
	for _, dir := range []string{"usr", "usr/local", "usr/local/share"} {
		info, err := os.Lstat(filepath.Join(root, dir))
		require.NoError(t, err)
		assert.Equal(t, os.ModeDir|0o755, info.Mode(), dir)
		assert.Equal(t, [2]uint32{0, 0}, owner(t, filepath.Join(root, dir)), dir)
	}
}
