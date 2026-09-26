//go:build kvm

package qemu_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/initramfs"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsock"
)

func TestAMachineBootsTheBuilderKernelAndAnswersOverVsock(t *testing.T) {
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

// builderMachine boots miso's pinned builder kernel with this test binary as
// its init. The kernel package is kept in miso's own cache, fetched once.
func builderMachine(t *testing.T) qemu.Machine {
	t.Helper()

	cache, err := os.UserCacheDir()
	require.NoError(t, err)
	kernel, err := builderkernel.Ready(t.Context(), download.Open(filepath.Join(cache, "miso", "bases"), http.DefaultClient))
	require.NoError(t, err)

	self, err := os.Executable()
	require.NoError(t, err)
	init, err := os.ReadFile(self)
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
		CommandLine: "console=ttyS0 panic=-1 MISO_GUEST=1",
		MemoryMiB:   512,
		CPUs:        1,
	}
}
