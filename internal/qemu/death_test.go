package qemu_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQEMUDiesWithMiso(t *testing.T) {
	// arrange
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid")
	t.Setenv("MISO_FAKE_QEMU", filepath.Join(dir, "arguments"))
	t.Setenv("MISO_FAKE_QEMU_HANG", "1")
	t.Setenv("MISO_FAKE_QEMU_PID", pidFile)
	t.Setenv("MISO_FAKE_MISO", "1")
	self, err := os.Executable()
	require.NoError(t, err)
	miso := exec.Command(self) //nolint:gosec // this test binary stands in for miso
	require.NoError(t, miso.Start())
	var qemu int
	require.Eventually(t, func() bool {
		qemu = pidIn(pidFile)

		return qemu != 0
	}, 10*time.Second, 10*time.Millisecond)
	t.Cleanup(func() { _ = syscall.Kill(qemu, syscall.SIGKILL) })

	// act
	require.NoError(t, miso.Process.Kill())
	_ = miso.Wait()

	// assert
	assert.Eventually(t, func() bool { return !running(qemu) }, 5*time.Second, 10*time.Millisecond)
}

// pidIn is the process ID written to the file, or 0 while there is none.
func pidIn(path string) int {
	written, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	pid, err := strconv.Atoi(string(written))
	if err != nil {
		return 0
	}

	return pid
}

// running is whether the process is there and not only waiting to be reaped.
func running(pid int) bool {
	stat, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return false
	}

	// the state follows the command name, which is in parentheses
	state := strings.Fields(string(stat[strings.LastIndexByte(string(stat), ')')+1:]))

	return len(state) > 0 && state[0] != "Z"
}
