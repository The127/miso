//go:build kvm

package main_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/initramfs"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsock"
)

func TestAnAgentBootedByTheBuilderKernelSaysWhyItDidNotStart(t *testing.T) {
	// arrange
	machine := builderMachine(t)
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
	err = ask(t, vm, protocol.Import{Key: "base", Digest: "sha256:0000000000000000000000000000000000000000000000000000000000000000"}, 30*time.Second)
	require.ErrorIs(t, err, protocol.ErrAgentFailed)
	assert.ErrorContains(t, err, "agent did not start")
	assert.ErrorContains(t, err, "miso-cache")
}

// ask sends the agent a request, dialling until it listens, the VM stops or
// the time is up.
func ask(t *testing.T, vm *qemu.VM, request protocol.Message, patience time.Duration) error {
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

		conn, err := vsock.Dial(vm.CID(), vsock.AgentPort)
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

// builderMachine boots miso's pinned builder kernel with the miso under test
// as its init, started as its agent. The kernel package is kept in miso's
// own cache, fetched once.
func builderMachine(t *testing.T) qemu.Machine {
	t.Helper()

	cache, err := os.UserCacheDir()
	require.NoError(t, err)
	kernel, err := builderkernel.Ready(t.Context(), download.Open(filepath.Join(cache, "miso", "bases"), http.DefaultClient))
	require.NoError(t, err)

	init, err := os.ReadFile(miso(t))
	require.NoError(t, err)

	dir := t.TempDir()
	image := filepath.Join(dir, "vmlinuz")
	require.NoError(t, os.WriteFile(image, kernel.Image, 0o600))
	var initrd bytes.Buffer
	require.NoError(t, initramfs.Write(&initrd, init, kernel.Modules))
	initramfsPath := filepath.Join(dir, "initramfs")
	require.NoError(t, os.WriteFile(initramfsPath, initrd.Bytes(), 0o600))

	return qemu.Machine{
		Kernel:      image,
		Initramfs:   initramfsPath,
		CommandLine: "console=ttyS0 panic=-1 -- agent",
		MemoryMiB:   512,
		CPUs:        1,
	}
}
