//go:build vmtest

package sandbox_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

func TestWhatARunPrintsGoesOut(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "echo hi"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "hi\n", out.String())
}

func TestWhatARunPrintsAsAnErrorGoesOut(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "echo hi >&2"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "hi\n", out.String())
}

func TestAFailedCommandAnswersItsExitCode(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "exit 3"}

	// act
	code, err := sandbox.Run(context.Background(), root, run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 3, code)
}

func TestACancelledRunStopsItsCommand(t *testing.T) {
	// arrange
	root := onBase(t)
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Second, cancel)
	run := protocol.Run{Command: "sleep 1000"}

	// act
	_, err := sandbox.Run(ctx, root, run, io.Discard)

	// assert
	assert.ErrorIs(t, err, context.Canceled)
}

// killWhenRunning kills the process with a command line with a signal once it
// runs.
func killWhenRunning(t *testing.T, commandLine string, signal syscall.Signal) {
	t.Helper()

	go func() {
		for {
			lines, _ := filepath.Glob("/proc/[0-9]*/cmdline")
			for _, line := range lines {
				found, _ := os.ReadFile(line)
				if string(found) == commandLine {
					pid, _ := strconv.Atoi(filepath.Base(filepath.Dir(line)))
					_ = syscall.Kill(pid, signal)

					return
				}
			}

			time.Sleep(10 * time.Millisecond)
		}
	}()
}

func TestARunKilledBySignalAnswersTheCodeAShellWould(t *testing.T) {
	// arrange
	root := onBase(t)
	killWhenRunning(t, "/bin/sh\x00-c\x00sleep 1000\x00", syscall.SIGKILL)
	run := protocol.Run{Command: "sleep 1000"}

	// act
	code, err := sandbox.Run(context.Background(), root, run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 137, code)
}
