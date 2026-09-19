//go:build vmtest

package layer_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/disk"
	"github.com/The127/miso/internal/layer"
)

func TestAFinishedLayerIsWholeAfterACrash(t *testing.T) {
	// arrange
	image := inMemory(t, "miso-test-crash")
	dir := mounted(t, image)
	store := layer.Open(dir)
	work, err := store.Begin("abc")
	require.NoError(t, err)
	written := bytes.Repeat([]byte("miso"), 1<<18)
	require.NoError(t, os.WriteFile(filepath.Join(work.Dir(), "hello"), written, 0o600))

	// act
	err = layer.Rename(work)

	// assert
	require.NoError(t, err)
	// the journal commits the rename, as it would a moment later on its own
	syncDir(t, dir)
	crashed := copied(t, image)
	got, err := os.ReadFile(filepath.Join(mounted(t, crashed), "abc", "hello"))
	require.NoError(t, err)
	assert.Equal(t, len(written), len(got))
}

func TestAFinishedLayerIsThereAfterACrashRightAfterIt(t *testing.T) {
	// arrange
	image := inMemory(t, "miso-test-crash")
	dir := mounted(t, image)
	work, err := layer.Open(dir).Begin("abc")
	require.NoError(t, err)

	// act
	err = work.Finish()

	// assert
	require.NoError(t, err)
	crashed := copied(t, image)
	assert.DirExists(t, filepath.Join(mounted(t, crashed), "abc"))
}

// inMemory copies the disk with a serial into a file on the tmpfs, where a
// copy of it is the disk as a crash would leave it.
func inMemory(t *testing.T, serial string) string {
	t.Helper()

	name, err := disk.BySerial(os.DirFS("/sys/block"), serial)
	require.NoError(t, err)

	return copied(t, filepath.Join("/dev", name))
}

// copied copies a file or disk into a new file on the tmpfs.
func copied(t *testing.T, path string) string {
	t.Helper()

	source, err := os.Open(path)
	require.NoError(t, err)

	defer func() { _ = source.Close() }()

	duplicate := filepath.Join(t.TempDir(), "disk.img")
	target, err := os.Create(duplicate)
	require.NoError(t, err)

	defer func() { _ = target.Close() }()

	_, err = io.Copy(target, source)
	require.NoError(t, err)

	return duplicate
}

// mounted mounts the ext4 in an image through a loop device of its own.
func mounted(t *testing.T, image string) string {
	t.Helper()

	control, err := os.Open("/dev/loop-control")
	require.NoError(t, err)

	defer func() { _ = control.Close() }()

	number, err := unix.IoctlRetInt(int(control.Fd()), unix.LOOP_CTL_GET_FREE)
	require.NoError(t, err)

	device, err := os.OpenFile(fmt.Sprintf("/dev/loop%d", number), os.O_RDWR, 0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = device.Close() })

	backing, err := os.OpenFile(image, os.O_RDWR, 0)
	require.NoError(t, err)

	defer func() { _ = backing.Close() }()

	require.NoError(t, unix.IoctlLoopConfigure(int(device.Fd()), &unix.LoopConfig{Fd: uint32(backing.Fd())})) //nolint:gosec // the kernel keeps file descriptors in an int
	t.Cleanup(func() { _ = unix.IoctlSetInt(int(device.Fd()), unix.LOOP_CLR_FD, 0) })

	dir := t.TempDir()
	require.NoError(t, syscall.Mount(device.Name(), dir, "ext4", 0, ""))
	t.Cleanup(func() { assert.NoError(t, syscall.Unmount(dir, 0)) })

	return dir
}

// syncDir has the file system commit what changed in a directory.
func syncDir(t *testing.T, dir string) {
	t.Helper()

	f, err := os.Open(dir)
	require.NoError(t, err)

	defer func() { _ = f.Close() }()

	require.NoError(t, f.Sync())
}
