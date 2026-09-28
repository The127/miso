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

func TestAFileLandsAtItsPathWithItsContent(t *testing.T) {
	// arrange
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "etc"), 0o700))

	// act
	err := place.Open(root).File("/etc/motd", 0o644, strings.NewReader("hello\n"))

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(root, "etc", "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(written))
}

func TestAFilesModeSurvivesTheUmask(t *testing.T) {
	// arrange
	root := t.TempDir()
	// the umask every machine might have, so the test fails anywhere
	// without the guard
	old := syscall.Umask(0o022)
	t.Cleanup(func() { syscall.Umask(old) })

	// act
	err := place.Open(root).File("/run.sh", 0o777, strings.NewReader("#!/bin/sh\n"))

	// assert
	require.NoError(t, err)
	info, err := os.Stat(filepath.Join(root, "run.sh"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o777), info.Mode().Perm())
}
