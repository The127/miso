package buildcontext_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func chmod(t *testing.T, dir string, name string, mode os.FileMode) {
	t.Helper()

	require.NoError(t, os.Chmod(filepath.Join(dir, name), mode))
}

func touch(t *testing.T, dir string, name string, when time.Time) {
	t.Helper()

	require.NoError(t, os.Chtimes(filepath.Join(dir, name), when, when))
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

func TestAChangedFileModeChangesTheDigest(t *testing.T) {
	// arrange
	plain := t.TempDir()
	write(t, plain, "run.sh", "#!/bin/sh")
	chmod(t, plain, "run.sh", 0o644)
	executable := t.TempDir()
	write(t, executable, "run.sh", "#!/bin/sh")
	chmod(t, executable, "run.sh", 0o755)

	// act
	plainDigest, plainErr := open(t, plain).Digest("run.sh")
	executableDigest, executableErr := open(t, executable).Digest("run.sh")

	// assert
	require.NoError(t, plainErr)
	require.NoError(t, executableErr)
	assert.NotEqual(t, plainDigest, executableDigest)
}

func TestAChangedModificationTimeKeepsTheDigest(t *testing.T) {
	// arrange
	old := t.TempDir()
	write(t, old, "motd", "hello")
	touch(t, old, "motd", time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC))
	recent := t.TempDir()
	write(t, recent, "motd", "hello")
	touch(t, recent, "motd", time.Date(2026, time.September, 17, 0, 0, 0, 0, time.UTC))

	// act
	oldDigest, oldErr := open(t, old).Digest("motd")
	recentDigest, recentErr := open(t, recent).Digest("motd")

	// assert
	require.NoError(t, oldErr)
	require.NoError(t, recentErr)
	assert.Equal(t, oldDigest, recentDigest)
}
