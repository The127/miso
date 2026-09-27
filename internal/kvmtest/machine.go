package kvmtest

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/initramfs"
	"github.com/The127/miso/internal/qemu"
)

// Machine boots miso's pinned builder kernel with init as its init and the
// command line given. The kernel package is kept in miso's own cache,
// fetched once.
func Machine(t *testing.T, init []byte, commandLine string) qemu.Machine {
	t.Helper()

	cache, err := os.UserCacheDir()
	require.NoError(t, err)
	kernel, err := builderkernel.Ready(t.Context(), download.Open(filepath.Join(cache, "miso", "bases"), http.DefaultClient))
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
		CommandLine: commandLine,
		MemoryMiB:   512,
		CPUs:        1,
	}
}
