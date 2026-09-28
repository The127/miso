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
