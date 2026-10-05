package qemu_test

import (
	"bytes"
	"context"
	"io"
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
	driver.OpenVsock = hostVsock
	machine := qemu.Machine{Boot: qemu.Kernel{Image: "/k/vmlinuz"}, MemoryMiB: 512, CPUs: 1}

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
	assert.Equal(t, slices.Concat(want, accelerated, qemu.Vsock(vm.CID(), 3, qemu.Machine{})), strings.Split(string(written), "\n"))
}

func TestAnArm64MachineRunsOnQEMUsAarch64Binary(t *testing.T) {
	// arrange
	driver, recorded := fakeDriver(t)
	path := t.TempDir()
	require.NoError(t, os.Symlink(driver.Binary, filepath.Join(path, "qemu-system-aarch64")))
	t.Setenv("PATH", path)
	driver.Binary = ""
	machine := qemu.Machine{Arch: "arm64", Boot: qemu.Kernel{Image: "/k/Image"}, MemoryMiB: 512, CPUs: 1}

	// act
	vm, err := driver.Start(t.Context(), machine)
	require.NoError(t, err)
	<-vm.Done()

	// assert
	assert.FileExists(t, recorded)
}

func TestAnAmd64MachineRunsOnQEMUsX8664Binary(t *testing.T) {
	// arrange
	driver, recorded := fakeDriver(t)
	path := t.TempDir()
	require.NoError(t, os.Symlink(driver.Binary, filepath.Join(path, "qemu-system-x86_64")))
	t.Setenv("PATH", path)
	driver.Binary = ""
	machine := qemu.Machine{Boot: qemu.Kernel{Image: "/k/vmlinuz"}, MemoryMiB: 512, CPUs: 1}

	// act
	vm, err := driver.Start(t.Context(), machine)
	require.NoError(t, err)
	<-vm.Done()

	// assert
	assert.FileExists(t, recorded)
}

func TestAMicrovmRunsOnQEMUsX8664Binary(t *testing.T) {
	// arrange
	driver, recorded := fakeDriver(t)
	path := t.TempDir()
	require.NoError(t, os.Symlink(driver.Binary, filepath.Join(path, "qemu-system-x86_64")))
	t.Setenv("PATH", path)
	driver.Binary = ""
	machine := qemu.Machine{Microvm: true, Boot: qemu.Kernel{Image: "/k/vmlinux"}, MemoryMiB: 512, CPUs: 1}

	// act
	vm, err := driver.Start(t.Context(), machine)
	require.NoError(t, err)
	<-vm.Done()

	// assert
	assert.FileExists(t, recorded)
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
	driver.OpenVsock = hostVsock
	t.Setenv("MISO_FAKE_QEMU_HANG", "1")
	vm, err := driver.Start(t.Context(), qemu.Machine{})
	require.NoError(t, err)
	other, err := hostVsock()
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
	machine := qemu.Machine{Boot: qemu.Kernel{Image: "/k/vmlinuz"}, MemoryMiB: 512, CPUs: 1, Temp: "/c/boot"}

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
	machine := qemu.Machine{Boot: qemu.Kernel{Image: "/k/vmlinuz"}, MemoryMiB: 512, CPUs: 1}

	// act
	vm, err := driver.Start(t.Context(), machine)
	require.NoError(t, err)
	<-vm.Done()

	// assert
	temp, err := os.ReadFile(told)
	require.NoError(t, err)
	assert.Equal(t, "/host/tmp", string(temp))
}

// noVsock is a host's vsock device that is not there.
func noVsock() (*os.File, error) {
	return nil, fs.ErrNotExist
}

func TestAHostWithoutVsockGivesTheMachineAVirtioPortForItsAgent(t *testing.T) {
	// arrange
	driver, recorded := fakeDriver(t)
	driver.OpenVsock = noVsock

	// act
	vm, err := driver.Start(t.Context(), qemu.Machine{})

	// assert
	require.NoError(t, err)
	<-vm.Done()
	written, err := os.ReadFile(recorded)
	require.NoError(t, err)
	args := strings.Split(string(written), "\n")
	assert.Equal(t, "socket,id=agent,fd=3", valueOf(t, args, "-chardev"))
	assert.Contains(t, args, "virtserialport,chardev=agent,name=miso")
}

func TestAHostWithoutVsockIsToldWhyItsMachineHasAVirtioPort(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	var told error
	driver.WithoutVsock = func(why error) { told = why }
	driver.OpenVsock = noVsock

	// act
	vm, err := driver.Start(t.Context(), qemu.Machine{})

	// assert
	require.NoError(t, err)
	<-vm.Done()
	assert.ErrorIs(t, told, fs.ErrNotExist)
}

func TestTheHostsEndOfAVirtioPortReachesTheMachine(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	t.Setenv("MISO_FAKE_QEMU_ECHO", "1")
	driver.OpenVsock = noVsock
	vm, err := driver.Start(t.Context(), qemu.Machine{})
	require.NoError(t, err)
	t.Cleanup(func() { <-vm.Done() })
	port := vm.Port()

	// act
	_, err = port.Write([]byte("hello"))
	require.NoError(t, err)
	back := make([]byte, 5)
	_, err = io.ReadFull(port, back)
	_ = port.Close()

	// assert
	require.NoError(t, err)
	assert.Equal(t, "hello", string(back))
}

func TestAStoppedMachineLetsGoOfItsVirtioPort(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	driver.OpenVsock = noVsock
	vm, err := driver.Start(t.Context(), qemu.Machine{})
	require.NoError(t, err)

	// act
	<-vm.Done()

	// assert
	_, err = vm.Port().Write([]byte("hello"))
	assert.ErrorIs(t, err, os.ErrClosed)
}

func TestQEMUHoldsNothingMisosCallerLeftOpen(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	held := filepath.Join(t.TempDir(), "held")
	t.Setenv("MISO_FAKE_QEMU_FDS", held)
	left := filepath.Join(t.TempDir(), "left open by miso's caller")
	require.NoError(t, os.WriteFile(left, nil, 0o600))
	file, err := os.Open(left)
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })
	// what a caller leaves open is not close-on-exec, as a dup is not
	fd, err := unix.Dup(int(file.Fd()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = unix.Close(fd) })
	machine := qemu.Machine{Boot: qemu.Kernel{Image: "/k/vmlinuz"}, MemoryMiB: 512, CPUs: 1}

	// act
	vm, err := driver.Start(t.Context(), machine)
	require.NoError(t, err)
	<-vm.Done()

	// assert
	targets, err := os.ReadFile(held)
	require.NoError(t, err)
	require.NotEmpty(t, targets)
	assert.NotContains(t, strings.Split(string(targets), "\n"), left)
}

func TestADriverClaimsItsCIDOnTheDeviceItIsHanded(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	driver.OpenVsock = func() (*os.File, error) { return os.Open(os.DevNull) }
	var told error
	driver.WithoutVsock = func(why error) { told = why }

	// act
	vm, err := driver.Start(t.Context(), qemu.Machine{})

	// assert
	require.NoError(t, err)
	<-vm.Done()
	assert.ErrorIs(t, told, unix.ENOTTY)
}

func TestADriverHandedNoVsockDeviceGivesItsMachineAVirtioPort(t *testing.T) {
	// arrange
	driver, _ := fakeDriver(t)
	var told error
	driver.WithoutVsock = func(why error) { told = why }

	// act
	vm, err := driver.Start(t.Context(), qemu.Machine{})

	// assert
	require.NoError(t, err)
	<-vm.Done()
	assert.NotNil(t, vm.Port())
	assert.ErrorIs(t, told, qemu.ErrNoVsockDevice)
}
