package builder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

func TestTheBuildersCacheDiskIsWritableUnderTheCacheSerial(t *testing.T) {
	// act
	machine := builder.Build{Cache: "/cache/disk.img"}.Machine()

	// assert
	assert.Equal(t, []qemu.Disk{{Path: "/cache/disk.img", Format: "raw", Serial: protocol.CacheSerial}}, machine.Disks)
}

func TestEachBaseImageIsAttachedReadOnlyUnderItsSerial(t *testing.T) {
	// arrange
	digest := "sha256:6e1f3a0c9b2d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789ab"
	requests := []protocol.Message{
		protocol.Import{Key: "base", Digest: digest},
		protocol.Run{Key: "step", Layers: []string{"base"}, Command: "true"},
	}
	blob := func(digest string) string { return "/bases/" + digest }

	// act
	machine := builder.Build{Cache: "/cache/disk.img", Requests: requests, Blob: blob}.Machine()

	// assert
	require.Len(t, machine.Disks, 2)
	assert.Equal(t, qemu.Disk{Path: "/bases/" + digest, Format: "qcow2", Serial: protocol.Serial(digest), ReadOnly: true}, machine.Disks[1])
}

func TestTheBuilderHasTheCardItIsGiven(t *testing.T) {
	// arrange
	card, _ := builder.Network(builder.Resolving{IPv4: true})

	// act
	machine := builder.Build{Card: card}.Machine()

	// assert
	assert.Equal(t, &card, machine.Card)
}

func TestTheBuilderBootsMisoAsItsAgent(t *testing.T) {
	// arrange
	boot := builder.Boot{Kernel: "/boot/vmlinuz", Initramfs: "/boot/initramfs"}

	// act
	machine := builder.Build{Boot: boot}.Machine()

	// assert
	assert.Equal(t, "/boot/vmlinuz", machine.Kernel)
	assert.Equal(t, "/boot/initramfs", machine.Initramfs)
	assert.Equal(t, "console=ttyS0 panic=-1 -- agent", machine.CommandLine)
}
