//go:build vmtest

package sandbox_test

import (
	"bytes"
	"context"
	"fmt"
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

func TestARunSeesTheEnvironmentOfTheBuildFileAndNotTheAgents(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Env: []string{"GREETING=hi"}, Command: "env"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "GREETING=hi\n")
	assert.NotContains(t, out.String(), "MISO_VMTEST_BASE_DIGEST")
}

func TestARunWithoutEnvironmentFindsCommandsOnTheUsualPath(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "env"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\n")
	assert.NotContains(t, out.String(), "MISO_VMTEST_BASE_DIGEST")
}

func TestARunIsAtHomeInRoot(t *testing.T) {
	// arrange
	root := onBase(t)
	run := protocol.Run{Command: "env"}
	var out bytes.Buffer

	// act
	_, err := sandbox.Run(context.Background(), root, run, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "HOME=/root\n")
}

func TestAProcessARunLeavesBehindDoesNotOutliveIt(t *testing.T) {
	// arrange
	root := onBase(t)
	// not &, which needs a /dev/null, and the second sleep lets the first one
	// start before the shell is gone
	run := protocol.Run{Command: "setsid -f sleep 1000; sleep 1"}

	// act
	code, err := sandbox.Run(context.Background(), root, run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Empty(t, others(t))
}

// others are the command lines of the processes other than the test, which
// in the VM is alone apart from the kernel's threads, and those have none.
func others(t *testing.T) []string {
	t.Helper()

	lines, err := filepath.Glob("/proc/[0-9]*/cmdline")
	require.NoError(t, err)
	found := []string{}
	for _, line := range lines {
		commandLine, _ := os.ReadFile(line)
		if len(commandLine) > 0 && filepath.Dir(line) != fmt.Sprintf("/proc/%d", os.Getpid()) {
			found = append(found, string(commandLine))
		}
	}

	return found
}
