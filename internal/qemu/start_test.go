package qemu_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/qemu"
)

func TestQEMURunsWithTheMachinesArgumentsAndItsVsockDevice(t *testing.T) {
	// arrange
	recorded := filepath.Join(t.TempDir(), "arguments")
	t.Setenv("MISO_FAKE_QEMU", recorded)
	self, err := os.Executable()
	require.NoError(t, err)
	driver := qemu.Driver{Binary: self}
	machine := qemu.Machine{Kernel: "/k/vmlinuz", MemoryMiB: 512, CPUs: 1}

	// act
	vm, err := driver.Start(t.Context(), machine)
	require.NoError(t, err)
	<-vm.Done()

	// assert
	want, err := qemu.Arguments(machine)
	require.NoError(t, err)
	written, err := os.ReadFile(recorded)
	require.NoError(t, err)
	assert.Equal(t, append(want, qemu.Vsock(vm.CID(), 3)...), strings.Split(string(written), "\n"))
}

func TestACancelledMachineIsStopped(t *testing.T) {
	// arrange
	t.Setenv("MISO_FAKE_QEMU", filepath.Join(t.TempDir(), "arguments"))
	t.Setenv("MISO_FAKE_QEMU_HANG", "1")
	self, err := os.Executable()
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	vm, err := qemu.Driver{Binary: self}.Start(ctx, qemu.Machine{})
	require.NoError(t, err)

	// act
	cancel()

	// assert
	select {
	case <-vm.Done():
	case <-time.After(10 * time.Second):
		assert.Fail(t, "still running after cancel")
	}
}

func TestAMachinesCIDStaysHeldWhileQEMURuns(t *testing.T) {
	// arrange
	t.Setenv("MISO_FAKE_QEMU", filepath.Join(t.TempDir(), "arguments"))
	t.Setenv("MISO_FAKE_QEMU_HANG", "1")
	self, err := os.Executable()
	require.NoError(t, err)
	vm, err := qemu.Driver{Binary: self}.Start(t.Context(), qemu.Machine{})
	require.NoError(t, err)
	other, err := qemu.OpenVsock("/dev/vhost-vsock")
	require.NoError(t, err)
	t.Cleanup(func() { _ = other.Close() })

	// act
	err = qemu.TakeCID(other, vm.CID())

	// assert
	assert.ErrorIs(t, err, unix.EADDRINUSE)
}

func TestAMachinesConsoleReachesTheWriterItNames(t *testing.T) {
	// arrange
	t.Setenv("MISO_FAKE_QEMU", filepath.Join(t.TempDir(), "arguments"))
	self, err := os.Executable()
	require.NoError(t, err)
	var console bytes.Buffer

	// act
	vm, err := qemu.Driver{Binary: self}.Start(t.Context(), qemu.Machine{Console: &console})
	require.NoError(t, err)
	<-vm.Done()

	// assert
	assert.Equal(t, "fake QEMU console\n", console.String())
}

func TestAMachineQEMUFailsToRunSaysWhy(t *testing.T) {
	// arrange
	t.Setenv("MISO_FAKE_QEMU", filepath.Join(t.TempDir(), "arguments"))
	t.Setenv("MISO_FAKE_QEMU_FAIL", "Could not access KVM kernel module")
	self, err := os.Executable()
	require.NoError(t, err)

	// act
	vm, err := qemu.Driver{Binary: self}.Start(t.Context(), qemu.Machine{})
	require.NoError(t, err)
	<-vm.Done()

	// assert
	assert.ErrorContains(t, vm.Err(), "Could not access KVM kernel module")
}
