//go:build vmtest

package tree_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/tree"
)

func TestDevicesAndFifosStayWhatTheyAre(t *testing.T) {
	// arrange
	source := t.TempDir()
	target := t.TempDir()
	require.NoError(t, unix.Mknod(filepath.Join(source, "block"), unix.S_IFBLK|0o600, int(unix.Mkdev(8, 0)))) //nolint:gosec // the kernel keeps device numbers in 32 bits
	require.NoError(t, unix.Mknod(filepath.Join(source, "char"), unix.S_IFCHR|0o666, int(unix.Mkdev(1, 3))))  //nolint:gosec // the kernel keeps device numbers in 32 bits
	require.NoError(t, unix.Mkfifo(filepath.Join(source, "fifo"), 0o600))

	// act
	err := tree.Copy(source, target)

	// assert
	require.NoError(t, err)

	for _, name := range []string{"block", "char", "fifo"} {
		want := special(t, filepath.Join(source, name))
		got := special(t, filepath.Join(target, name))
		assert.Equal(t, want, got, name)
	}
}

// special is the kind of file a path is, with its mode and device number.
func special(t *testing.T, name string) [2]uint64 {
	t.Helper()

	info, err := os.Lstat(name)
	require.NoError(t, err)

	stat, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)

	return [2]uint64{uint64(stat.Mode), stat.Rdev}
}
