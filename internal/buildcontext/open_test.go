package buildcontext_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/buildcontext"
)

func TestAFileSwappedForAPipeIsRefusedInsteadOfWaitedFor(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, syscall.Mkfifo(filepath.Join(dir, "motd"), 0o600))
	looked, err := os.Lstat(filepath.Join(dir, "motd"))
	require.NoError(t, err)
	context := open(t, dir)
	opened := make(chan error, 1)

	// act
	go func() {
		file, err := buildcontext.OpenFile(context, "motd", looked)
		if err == nil {
			_ = file.Close()
		}

		opened <- err
	}()

	// assert
	select {
	case err := <-opened:
		assert.ErrorIs(t, err, buildcontext.ErrSpecialFile)
	case <-time.After(time.Second):
		t.Fatal("opening a pipe waited for a writer")
	}
}

func TestAFileSwappedForALinkAfterItWasLookedAtIsRefused(t *testing.T) {
	// arrange
	dir := t.TempDir()
	write(t, dir, "motd", "looked at")
	looked, err := os.Lstat(filepath.Join(dir, "motd"))
	require.NoError(t, err)
	write(t, dir, "real", "not the file looked at")
	require.NoError(t, os.Remove(filepath.Join(dir, "motd")))
	symlink(t, dir, "motd", "real")
	context := open(t, dir)

	// act
	file, err := buildcontext.OpenFile(context, "motd", looked)
	if err == nil {
		_ = file.Close()
	}

	// assert
	assert.ErrorIs(t, err, buildcontext.ErrSwapped)
}
