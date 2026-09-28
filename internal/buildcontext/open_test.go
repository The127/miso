package buildcontext_test

import (
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
	context := open(t, dir)
	opened := make(chan error, 1)

	// act
	go func() {
		file, err := buildcontext.OpenFile(context, "motd")
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
