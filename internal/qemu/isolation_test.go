//go:build kvm

package qemu_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/kvmtest"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsock"
	"github.com/The127/miso/internal/vsockns"
)

// private is a vsock namespace held for the test, which is skipped on a host
// that cannot keep vsock private.
func private(t *testing.T) *vsockns.Namespace {
	t.Helper()

	namespace, err := vsockns.Open()
	if errors.Is(err, vsockns.ErrNotPrivate) {
		t.Skip("this host cannot keep vsock private")
	}

	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, namespace.Close()) })

	return namespace
}

// guestIn is a VM with this binary as its guest, run on a device of the
// namespace.
func guestIn(t *testing.T, namespace *vsockns.Namespace) *qemu.VM {
	t.Helper()

	self, err := os.Executable()
	require.NoError(t, err)
	init, err := os.ReadFile(self)
	require.NoError(t, err)
	machine := kvmtest.Machine(t, init, "console=ttyS0 panic=-1 MISO_GUEST=1")
	var console bytes.Buffer
	machine.Console = &console
	vm, err := qemu.Driver{Binary: "qemu-system-x86_64", OpenVsock: namespace.Device}.Start(t.Context(), machine)
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
	namespace := private(t)
	vm := guestIn(t, namespace)
	inside := func() (io.ReadWriteCloser, error) {
		socket, err := namespace.Socket()
		if err != nil {
			return nil, err
		}

		defer func() { _ = socket.Close() }()

		return vsock.DialOn(socket, vm.CID(), guestPort)
	}
	said, err := answer(vm, inside, 30*time.Second)
	require.NoError(t, err)
	require.Equal(t, "miso\n", said, "the guest listens inside")

	// act
	_, err = vsock.Dial(vm.CID(), guestPort)

	// assert
	assert.Error(t, err)
}
