//go:build vmtest

package tree_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/tree"
)

// onto lands everything a copy takes at one path of the image.
func onto(path string) tree.Land {
	return func(string, string, bool) (string, error) { return path, nil }
}

func TestAFileOfAStageLandsAtTheDestination(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "etc"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "etc", "motd"), []byte("hello\n"), 0o600))
	land := onto("/motd")

	// act
	err := tree.Into(stage, place.Open(image), []string{"/etc/motd"}, land)

	// assert
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(image, "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(got))
}

func TestAFileOfAStageKeepsItsOwnerModeAndTimes(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	tool := filepath.Join(stage, "tool")
	require.NoError(t, os.WriteFile(tool, []byte("#!/bin/sh\n"), 0o600))
	require.NoError(t, os.Chown(tool, 1234, 5678))
	require.NoError(t, os.Chmod(tool, os.ModeSetuid|0o755))
	then := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	require.NoError(t, os.Chtimes(tool, then, then))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/tool"}, onto("/tool"))

	// assert
	require.NoError(t, err)
	info, err := os.Stat(filepath.Join(image, "tool"))
	require.NoError(t, err)
	stat, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	assert.Equal(t, uint32(1234), stat.Uid)
	assert.Equal(t, uint32(5678), stat.Gid)
	assert.Equal(t, os.ModeSetuid|0o755, info.Mode())
	assert.True(t, then.Equal(info.ModTime()), "modified %s", info.ModTime())
}

func TestAFileCapabilityOfAStageOutlivesItsOwnerChange(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	ping := filepath.Join(stage, "ping")
	require.NoError(t, os.WriteFile(ping, nil, 0o600))
	require.NoError(t, os.Chown(ping, 1234, 1234))
	require.NoError(t, unix.Lsetxattr(ping, "security.capability", netRaw, 0))
	require.NoError(t, unix.Lsetxattr(ping, "user.note", []byte("file"), 0))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/ping"}, onto("/ping"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, netRaw, xattr(t, filepath.Join(image, "ping"), "security.capability"))
	assert.Equal(t, "file", string(xattr(t, filepath.Join(image, "ping"), "user.note")))
}

// below lands what a copy takes below one directory of the image.
func below(dir string) tree.Land {
	return func(_, path string, _ bool) (string, error) { return filepath.Join(dir, path), nil }
}

func TestADirectoryOfAStageCopiesWhatIsInIt(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "etc", "sub"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "etc", "motd"), []byte("hello\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "etc", "sub", "issue"), []byte("welcome\n"), 0o600))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/etc"}, below("/copied"))

	// assert
	require.NoError(t, err)
	motd, err := os.ReadFile(filepath.Join(image, "copied", "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(motd))
	issue, err := os.ReadFile(filepath.Join(image, "copied", "sub", "issue"))
	require.NoError(t, err)
	assert.Equal(t, "welcome\n", string(issue))
}

func TestADirectoryOfAStageKeepsItsOwnerModeAndTimesPastItsChildren(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	dir := filepath.Join(stage, "d")
	require.NoError(t, os.Mkdir(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f"), nil, 0o600))
	require.NoError(t, os.Chown(dir, 1234, 5678))
	require.NoError(t, os.Chmod(dir, os.ModeSticky|0o750))
	then := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	require.NoError(t, os.Chtimes(dir, then, then))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/d"}, below("/copied"))

	// assert
	require.NoError(t, err)
	info, err := os.Stat(filepath.Join(image, "copied"))
	require.NoError(t, err)
	stat, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	assert.Equal(t, uint32(1234), stat.Uid)
	assert.Equal(t, uint32(5678), stat.Gid)
	assert.Equal(t, os.ModeDir|os.ModeSticky|0o750, info.Mode())
	assert.True(t, then.Equal(info.ModTime()), "modified %s", info.ModTime())
}

func TestADirectoryOfTheImageKeepsItsOwnWhenAStageIsCopiedIntoIt(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(stage, "d"), 0o750))
	require.NoError(t, os.Chown(filepath.Join(stage, "d"), 1234, 5678))
	require.NoError(t, os.Mkdir(filepath.Join(image, "etc"), 0o700))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/d"}, below("/etc"))

	// assert
	require.NoError(t, err)
	info, err := os.Stat(filepath.Join(image, "etc"))
	require.NoError(t, err)
	stat, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	assert.Equal(t, uint32(0), stat.Uid)
	assert.Equal(t, os.ModeDir|0o700, info.Mode())
}
