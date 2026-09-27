package builder

import (
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

// Machine is the builder VM a build boots from boot, with miso as its agent,
// its cache disk at a path, a disk for each base image the requests import,
// found by blob, and the card its runs reach out through.
func Machine(boot Boot, cache string, requests []protocol.Message, blob func(digest string) string, card qemu.Card) qemu.Machine {
	// the cache disk is an ext4 image as it is, which QEMU must never guess
	disks := []qemu.Disk{{Path: cache, Format: "raw", Serial: protocol.CacheSerial}}
	for _, request := range requests {
		if request, isImport := request.(protocol.Import); isImport {
			// every base image miso knows is a qcow2 image
			disks = append(disks, qemu.Disk{Path: blob(request.Digest), Format: "qcow2", Serial: protocol.Serial(request.Digest), ReadOnly: true})
		}
	}

	return qemu.Machine{
		Kernel:    boot.Kernel,
		Initramfs: boot.Initramfs,
		// the kernel hands init what follows --, and init is miso
		CommandLine: "console=ttyS0 panic=-1 -- agent",
		Disks:       disks,
		Card:        &card,
	}
}
