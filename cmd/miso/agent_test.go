//go:build kvm

package main_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/kvmtest"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/reach"
	"github.com/The127/miso/internal/vsockns/vsocknstest"
)

func TestAnAgentBootedByTheBuilderKernelWithoutVsockAnswersOverItsVirtioPort(t *testing.T) {
	// arrange
	init, err := os.ReadFile(miso(t))
	require.NoError(t, err)
	machine := kvmtest.Machine(t, init, "console=ttyS0 panic=-1 -- agent")
	var console bytes.Buffer
	machine.Console = &console
	driver := qemu.Driver{Binary: "qemu-system-x86_64", OpenVsock: func() (*os.File, error) { return nil, fs.ErrNotExist }}

	// act
	vm, err := driver.Start(t.Context(), machine)
	require.NoError(t, err)
	t.Cleanup(func() {
		// QEMU writes to the console until it is gone
		<-vm.Done()
		t.Logf("QEMU: %v, its console:\n%s", vm.Err(), console.String())
	})

	// assert
	require.NotNil(t, vm.Port(), "reach falls back to vsock without a port")
	dial, err := reach.Agent(vm, nil)
	require.NoError(t, err)
	err = ask(t, vm, dial, protocol.Import{Key: "base", Digest: "sha256:0000000000000000000000000000000000000000000000000000000000000000"}, 30*time.Second)
	require.ErrorIs(t, err, protocol.ErrAgentFailed)
	assert.ErrorContains(t, err, "agent did not start")
	assert.ErrorContains(t, err, protocol.CacheSerial)
}

func TestAnAgentBootedByTheBuilderKernelInAVsockNamespaceIsReachedThroughIt(t *testing.T) {
	// arrange
	namespace := vsocknstest.Private(t)
	init, err := os.ReadFile(miso(t))
	require.NoError(t, err)
	machine := kvmtest.Machine(t, init, "console=ttyS0 panic=-1 -- agent")
	var console bytes.Buffer
	machine.Console = &console
	vm, err := qemu.Driver{Binary: "qemu-system-x86_64", OpenVsock: namespace.Device}.Start(t.Context(), machine)
	require.NoError(t, err)
	t.Cleanup(func() {
		// QEMU writes to the console until it is gone
		<-vm.Done()
		t.Logf("QEMU: %v, its console:\n%s", vm.Err(), console.String())
	})

	// act
	dial, err := reach.Agent(vm, namespace.Socket)

	// assert
	require.NoError(t, err)
	err = ask(t, vm, dial, protocol.Import{Key: "base", Digest: "sha256:0000000000000000000000000000000000000000000000000000000000000000"}, 30*time.Second)
	require.ErrorIs(t, err, protocol.ErrAgentFailed)
	assert.ErrorContains(t, err, "agent did not start")
}

// ask sends the agent a request on a connection the dial makes, dialling
// until it listens, the VM stops or the time is up.
func ask(t *testing.T, vm *qemu.VM, dial func() (io.ReadWriteCloser, error), request protocol.Message, patience time.Duration) error {
	t.Helper()

	deadline := time.After(patience)
	for {
		select {
		case <-vm.Done():
			return errors.New("the VM stopped before it answered, its console says why")
		case <-deadline:
			return fmt.Errorf("no answer after %s", patience)
		default:
		}

		conn, err := dial()
		if err != nil {
			time.Sleep(100 * time.Millisecond)

			continue
		}

		defer func() { _ = conn.Close() }()

		return protocol.New(agentName(t), conn, conn).Ask(request, io.Discard)
	}
}

// agentName is how the miso under test names itself to its agent.
func agentName(t *testing.T) string {
	t.Helper()

	version, err := exec.Command(miso(t), "version").Output() //nolint:gosec // the test names the binary
	require.NoError(t, err)

	return "miso " + strings.TrimSpace(string(version))
}

// miso is the miso binary under test, which MISO names, see just test-kvm.
func miso(t *testing.T) string {
	t.Helper()

	binary := os.Getenv("MISO")
	require.NotEmpty(t, binary, "MISO names the miso binary under test, see just test-kvm")

	return binary
}
