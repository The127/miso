package buildcontext_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func write(t *testing.T, dir string, name string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}

func mkdir(t *testing.T, dir string, name string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, name), 0o750))
}

func symlink(t *testing.T, dir string, name string, target string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o750))
	require.NoError(t, os.Symlink(target, filepath.Join(dir, name)))
}

func chmod(t *testing.T, dir string, name string, mode os.FileMode) {
	t.Helper()

	require.NoError(t, os.Chmod(filepath.Join(dir, name), mode))

	// some file systems keep no modes, and the kernel drops a setgid bit
	// for a group the caller is not in. A test about modes means nothing then
	info, err := os.Lstat(filepath.Join(dir, name))
	require.NoError(t, err)
	require.Equal(t, mode, info.Mode()&(os.ModePerm|os.ModeSetuid|os.ModeSetgid|os.ModeSticky), "the file system did not keep the mode")
}

func touch(t *testing.T, dir string, name string, when time.Time) {
	t.Helper()

	require.NoError(t, os.Chtimes(filepath.Join(dir, name), when, when))
}
