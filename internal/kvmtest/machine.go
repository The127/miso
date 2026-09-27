package kvmtest

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
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

	boot, err := builder.WriteBoot(t.TempDir(), kernel, init)
	require.NoError(t, err)

	return qemu.Machine{
		Kernel:      boot.Kernel,
		Initramfs:   boot.Initramfs,
		CommandLine: commandLine,
		MemoryMiB:   512,
		CPUs:        1,
	}
}
