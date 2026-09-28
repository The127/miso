package qemu_test

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
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

// startOn starts the machine of the driver as if the host's KVM device
// were kvm, and waits for it to stop at the end of the test.
func startOn(t *testing.T, driver qemu.Driver, kvm string) {
	t.Helper()

	vm, err := qemu.StartOn(t.Context(), driver, qemu.Machine{}, kvm)
	if err == nil {
		t.Cleanup(func() { <-vm.Done() })
	}
}

func TestAHostWithoutKVMIsToldWhyTheMachineRunsOnTCG(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	var told error
	driver.WithoutKVM = func(why error) { told = why }

	// act
	startOn(t, driver, filepath.Join(t.TempDir(), "kvm"))

	// assert
	assert.ErrorIs(t, told, fs.ErrNotExist)
}

func TestAHostWithoutKVMStartsTheMachineWithNobodyToTell(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)

	// act
	start := func() { startOn(t, driver, filepath.Join(t.TempDir(), "kvm")) }

	// assert
	assert.NotPanics(t, start)
}

func TestAMachineOnAHostWithoutKVMSaysSo(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)

	// act
	vm, err := qemu.StartOn(t.Context(), driver, qemu.Machine{}, filepath.Join(t.TempDir(), "kvm"))

	// assert
	require.NoError(t, err)
	<-vm.Done()
	assert.True(t, vm.WithoutKVM())
}

func TestQEMUKeepsItsTemporaryFilesWhereTheMachineSays(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	told := filepath.Join(t.TempDir(), "tmpdir")
	t.Setenv("MISO_FAKE_QEMU_TMPDIR", told)
	machine := qemu.Machine{Kernel: "/k/vmlinuz", MemoryMiB: 512, CPUs: 1, Temp: "/c/boot"}

	// act
	vm, err := driver.Start(t.Context(), machine)
	require.NoError(t, err)
	<-vm.Done()

	// assert
	temp, err := os.ReadFile(told)
	require.NoError(t, err)
	assert.Equal(t, "/c/boot", string(temp))
}

func TestAMachineWithoutATempDirLeavesQEMUTheHostsOwn(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	told := filepath.Join(t.TempDir(), "tmpdir")
	t.Setenv("MISO_FAKE_QEMU_TMPDIR", told)
	t.Setenv("TMPDIR", "/host/tmp")
	machine := qemu.Machine{Kernel: "/k/vmlinuz", MemoryMiB: 512, CPUs: 1}

	// act
	vm, err := driver.Start(t.Context(), machine)
	require.NoError(t, err)
	<-vm.Done()

	// assert
	temp, err := os.ReadFile(told)
	require.NoError(t, err)
	assert.Equal(t, "/host/tmp", string(temp))
}
