package qemu_test

import (
	"bytes"
	"context"
	"os"
	"slices"
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
	driver, recorded := fakeDriver(t)
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
	accelerated, _ := qemu.Accel("/dev/kvm")
	assert.Equal(t, slices.Concat(want, accelerated, qemu.Vsock(vm.CID(), 3)), strings.Split(string(written), "\n"))
}

func TestACancelledMachineIsStopped(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	t.Setenv("MISO_FAKE_QEMU_HANG", "1")
	ctx, cancel := context.WithCancel(t.Context())
	vm, err := driver.Start(ctx, qemu.Machine{})
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
	driver, _ := fakeDriver(t)
	t.Setenv("MISO_FAKE_QEMU_HANG", "1")
	vm, err := driver.Start(t.Context(), qemu.Machine{})
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
	driver, _ := fakeDriver(t)
	var console bytes.Buffer

	// act
	vm, err := driver.Start(t.Context(), qemu.Machine{Console: &console})
	require.NoError(t, err)
	<-vm.Done()

	// assert
	assert.Equal(t, "fake QEMU console\n", console.String())
}

func TestAMachineQEMUFailsToRunSaysWhy(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	t.Setenv("MISO_FAKE_QEMU_FAIL", "Could not access KVM kernel module")

	// act
	vm, err := driver.Start(t.Context(), qemu.Machine{})
	require.NoError(t, err)
	<-vm.Done()

	// assert
	assert.ErrorContains(t, vm.Err(), "Could not access KVM kernel module")
}

func TestAMachineQEMUStoppedWithoutAWordSaysOnlyHow(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	t.Setenv("MISO_FAKE_QEMU_HANG", "1")
	ctx, cancel := context.WithCancel(t.Context())
	vm, err := driver.Start(ctx, qemu.Machine{})
	require.NoError(t, err)

	// act
	cancel()
	<-vm.Done()

	// assert
	assert.EqualError(t, vm.Err(), "QEMU stopped: signal: killed")
}
