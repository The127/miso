package kvmtest

import (
	"net/http"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/qemu"
)

// Machine boots miso's pinned builder kernel for the host's architecture
// with init as its init. The command line given follows its console= word.
// The kernel package is kept in miso's own cache, fetched once.
func Machine(t *testing.T, init []byte, commandLine string) qemu.Machine {
	t.Helper()

	kernel, err := builderkernel.Ready(t.Context(), download.Open(basesDir(t), http.DefaultClient), runtime.GOARCH)
	require.NoError(t, err)

	boot, err := builder.WriteBoot(t.TempDir(), runtime.GOARCH, kernel, init)
	require.NoError(t, err)

	return qemu.Machine{
		Boot:      qemu.Kernel{Image: boot.Kernel, Initramfs: boot.Initramfs, CommandLine: "console=" + builder.Console(runtime.GOARCH) + " " + commandLine},
		MemoryMiB: 512,
		CPUs:      1,
		Arch:      runtime.GOARCH,
	}
}
