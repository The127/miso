//go:build kvm

package qemu_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/kvmtest"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsock"
)

func TestAMachineBootsTheBuilderKernelAndAnswersOverVsock(t *testing.T) {
	// arrange
	self, err := os.Executable()
	require.NoError(t, err)
	init, err := os.ReadFile(self)
	require.NoError(t, err)
	machine := kvmtest.Machine(t, init, "console=ttyS0 panic=-1 MISO_GUEST=1")
	var console bytes.Buffer
	machine.Console = &console

	// act
	vm, err := qemu.Driver{Binary: "qemu-system-x86_64"}.Start(t.Context(), machine)
	require.NoError(t, err)
	t.Cleanup(func() {
		// QEMU writes to the console until it is gone
		<-vm.Done()
		t.Logf("QEMU: %v, its console:\n%s", vm.Err(), console.String())
	})

	// assert
	said, err := answer(vm, 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "miso\n", said)
}

// answer is what the guest says over vsock, asked until it listens, the VM
// stops or the time is up.
func answer(vm *qemu.VM, patience time.Duration) (string, error) {
	deadline := time.After(patience)
	for {
		select {
		case <-vm.Done():
			return "", errors.New("the VM stopped before it answered, its console says why")
		case <-deadline:
			return "", fmt.Errorf("no answer after %s", patience)
		default:
		}

		conn, err := vsock.Dial(vm.CID(), guestPort)
		if err != nil {
			time.Sleep(100 * time.Millisecond)

			continue
		}

		said, err := io.ReadAll(conn)
		_ = conn.Close()

		return string(said), err
	}
}
