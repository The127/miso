package qemu_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestQEMUDiesWithMiso(t *testing.T) {
	// arrange
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid")
	driver, _ := fakeDriver(t)
	t.Setenv("MISO_FAKE_QEMU_HANG", "1")
	t.Setenv("MISO_FAKE_QEMU_PID", pidFile)
	t.Setenv("MISO_FAKE_MISO", "1")
	miso := exec.Command(driver.Binary) //nolint:gosec // this test binary stands in for miso
	require.NoError(t, miso.Start())
	var qemuPid int
	require.Eventually(t, func() bool {
		qemuPid = pidIn(pidFile)

		return qemuPid != 0
	}, 10*time.Second, 10*time.Millisecond)
	t.Cleanup(func() { _ = syscall.Kill(qemuPid, syscall.SIGKILL) })

	// act
	require.NoError(t, miso.Process.Kill())
	_ = miso.Wait()

	// assert
	assert.Eventually(t, func() bool { return !running(qemuPid) }, 5*time.Second, 10*time.Millisecond)
}

func TestQEMUOutlivesTheThreadThatStartedIt(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	t.Setenv("MISO_FAKE_QEMU_HANG", "1")
	type start struct {
		vm  *qemu.VM
		err error
	}
	started := make(chan start)

	// act
	go func() {
		// a goroutine that ends locked takes its thread with it
		runtime.LockOSThread()

		vm, err := driver.Start(t.Context(), qemu.Machine{})
		started <- start{vm, err}
	}()
	result := <-started
	require.NoError(t, result.err)

	// assert
	assert.Never(t, func() bool {
		select {
		case <-result.vm.Done():
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestACtrlCInMisosTerminalReachesOnlyMiso(t *testing.T) {
	// arrange
	pidFile := filepath.Join(t.TempDir(), "pid")
	driver, _ := fakeDriver(t)
	t.Setenv("MISO_FAKE_QEMU_HANG", "1")
	t.Setenv("MISO_FAKE_QEMU_PID", pidFile)

	// act
	_, err := driver.Start(t.Context(), qemu.Machine{})
	require.NoError(t, err)

	// assert
	var pid int
	require.Eventually(t, func() bool {
		pid = pidIn(pidFile)

		return pid != 0
	}, 10*time.Second, 10*time.Millisecond)
	group, err := syscall.Getpgid(pid)
	require.NoError(t, err)
	// the terminal sends Ctrl-C to its foreground process group
	assert.NotEqual(t, syscall.Getpgrp(), group)
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
