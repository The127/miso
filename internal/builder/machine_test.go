package builder_test

import (
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

func TestTheBuildersCacheDiskIsWritableUnderTheCacheSerial(t *testing.T) {
	// act
	machine, err := builder.Build{Cache: "/cache/disk.img"}.Machine()

	// assert
	require.NoError(t, err)
	assert.Equal(t, []qemu.Disk{{Path: "/cache/disk.img", Format: "raw", Serial: protocol.CacheSerial}}, machine.Disks)
}

func TestEachBaseImageIsAttachedReadOnlyUnderItsSerial(t *testing.T) {
	// arrange
	digest := "sha256:6e1f3a0c9b2d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789ab"
	requests := requested(
		protocol.Import{Key: "base", Digest: digest},
		protocol.Run{Key: "step", Layers: []string{"base"}, Command: "true"},
	)
	blob := func(digest string) string { return "/bases/" + digest }
	format := func(string) (string, error) { return "qcow2", nil }

	// act
	machine, err := builder.Build{Cache: "/cache/disk.img", Requests: requests, Blob: blob, Format: format}.Machine()

	// assert
	require.NoError(t, err)
	require.Len(t, machine.Disks, 2)
	assert.Equal(t, qemu.Disk{Path: "/bases/" + digest, Format: "qcow2", Serial: protocol.Serial(digest), Access: qemu.ReadOnly}, machine.Disks[1])
}

func TestABaseImagesDiskHasTheFormatItWasFetchedIn(t *testing.T) {
	// arrange
	requests := requested(protocol.Import{Key: "base", Digest: "sha256:6e1f3a0c9b2d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789ab"})
	blob := func(digest string) string { return "/bases/" + digest }
	format := func(string) (string, error) { return "raw", nil }

	// act
	machine, err := builder.Build{Requests: requests, Blob: blob, Format: format}.Machine()

	// assert
	require.NoError(t, err)
	require.Len(t, machine.Disks, 2)
	assert.Equal(t, "raw", machine.Disks[1].Format)
}

func TestABaseImageWhoseFormatIsUnknownFailsTheMachine(t *testing.T) {
	// arrange
	unknown := errors.New("no image fetched with that digest")
	requests := requested(protocol.Import{Key: "base", Digest: "sha256:6e1f3a0c9b2d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789ab"})
	blob := func(digest string) string { return "/bases/" + digest }
	format := func(string) (string, error) { return "", unknown }

	// act
	_, err := builder.Build{Requests: requests, Blob: blob, Format: format}.Machine()

	// assert
	assert.ErrorIs(t, err, unknown)
}

func TestTheBuilderHasTheCardItIsGiven(t *testing.T) {
	// arrange
	card, _ := builder.Network(builder.Resolving{IPv4: true})

	// act
	machine, err := builder.Build{Card: card}.Machine()

	// assert
	require.NoError(t, err)
	assert.Equal(t, &card, machine.Card)
}

func TestTheBuilderBootsMisoAsItsAgent(t *testing.T) {
	// arrange
	boot := builder.Boot{Kernel: "/boot/vmlinuz", Initramfs: "/boot/initramfs"}

	// act
	machine, err := builder.Build{Boot: boot}.Machine()

	// assert
	require.NoError(t, err)
	assert.Equal(t, "/boot/vmlinuz", machine.Kernel)
	assert.Equal(t, "/boot/initramfs", machine.Initramfs)
	assert.Equal(t, "console=ttyS0 panic=-1 -- agent", machine.CommandLine)
}

func TestTheBuilderHasFourGibibytesOfMemory(t *testing.T) {
	// act
	machine, err := builder.Build{}.Machine()

	// assert
	require.NoError(t, err)
	assert.Equal(t, 4096, machine.MemoryMiB)
}

func TestTheBuilderHasTheHostsCPUsUpToEight(t *testing.T) {
	// act
	machine, err := builder.Build{}.Machine()

	// assert
	require.NoError(t, err)
	assert.Equal(t, min(runtime.NumCPU(), 8), machine.CPUs)
}
