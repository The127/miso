//go:build kvm

package qemu_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/kvmtest"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsock"
	"github.com/The127/miso/internal/vsockns"
	"github.com/The127/miso/internal/vsockns/vsocknstest"
)

// guestIn is a VM with this binary as its guest, run on a device of the
// namespace.
func guestIn(t *testing.T, namespace *vsockns.Namespace, cmdline string) *qemu.VM {
	t.Helper()

	self, err := os.Executable()
	require.NoError(t, err)
	init, err := os.ReadFile(self)
	require.NoError(t, err)
	machine := kvmtest.Machine(t, init, "console=ttyS0 panic=-1 MISO_GUEST=1 "+cmdline)
	var console bytes.Buffer
	machine.Console = &console
	vm, err := qemu.Driver{OpenVsock: namespace.Device}.Start(t.Context(), machine)
	require.NoError(t, err)
	t.Cleanup(func() {
		// QEMU writes to the console until it is gone
		<-vm.Done()
		t.Logf("QEMU: %v, its console:\n%s", vm.Err(), console.String())
	})

	return vm
}

func TestAMachineOnADeviceOfAVsockNamespaceIsOutOfTheHostsReachWithKVM(t *testing.T) {
	// arrange
	namespace := vsocknstest.Private(t)
	vm := guestIn(t, namespace, "")
	said, err := answer(vm, insideDial(namespace, vm))
	require.NoError(t, err)
	require.Equal(t, "miso\n", said, "the guest listens inside")

	// act
	_, err = vsock.DialOn(plainSocket, vm.CID(), guestPort)

	// assert
	assert.Error(t, err)
}

// insideDial dials the guest through sockets of the namespace.
func insideDial(namespace *vsockns.Namespace, vm *qemu.VM) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) { return vsock.DialOn(namespace.Socket, vm.CID(), guestPort) }
}

func TestAMachineOnADeviceOfAVsockNamespaceCannotReachTheHostWithKVM(t *testing.T) {
	// arrange
	namespace := vsocknstest.Private(t)
	outside, err := vsock.Listen(unix.VMADDR_PORT_ANY)
	require.NoError(t, err)
	t.Cleanup(func() { _ = outside.Close() })

	// act
	vm := guestIn(t, namespace, fmt.Sprintf("MISO_GUEST_DIAL=%d", outside.Port()))

	// assert
	said, err := answer(vm, insideDial(namespace, vm))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(said, "not reached"), said)
}

func TestAMachineOnADeviceOfAVsockNamespaceReachesTheHostInsideItWithKVM(t *testing.T) {
	// arrange
	namespace := vsocknstest.Private(t)
	socket, err := namespace.Listen(unix.VMADDR_PORT_ANY)
	require.NoError(t, err)
	t.Cleanup(func() { _ = socket.Close() })
	inside, err := vsock.Listening(socket)
	require.NoError(t, err)
	t.Cleanup(func() { _ = inside.Close() })

	// act
	vm := guestIn(t, namespace, fmt.Sprintf("MISO_GUEST_DIAL=%d", inside.Port()))

	// assert
	said, err := answer(vm, insideDial(namespace, vm))
	require.NoError(t, err)
	assert.Equal(t, "reached\n", said)
}
