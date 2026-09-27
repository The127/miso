package builder

import (
	"runtime"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

// Build is what a build's builder VM is made from.
type Build struct {
	// the files it boots from, with miso as its agent
	Boot Boot

	// the path of its cache disk
	Cache string

	// what the agent is asked, whose imports each get a disk
	Requests []protocol.Message

	// where the base image with a digest is
	Blob func(digest string) string

	// the card its runs reach out through
	Card qemu.Card
}

// Machine is the builder VM of the build.
func (b Build) Machine() qemu.Machine {
	// the cache disk is an ext4 image as it is, which QEMU must never guess
	disks := []qemu.Disk{{Path: b.Cache, Format: "raw", Serial: protocol.CacheSerial}}
	for _, request := range b.Requests {
		if request, isImport := request.(protocol.Import); isImport {
			// every base image miso knows is a qcow2 image
			disks = append(disks, qemu.Disk{Path: b.Blob(request.Digest), Format: "qcow2", Serial: protocol.Serial(request.Digest), ReadOnly: true})
		}
	}

	return qemu.Machine{
		Kernel:    b.Boot.Kernel,
		Initramfs: b.Boot.Initramfs,
		// the kernel hands init what follows --, and init is miso
		CommandLine: "console=ttyS0 panic=-1 -- agent",
		// room for a package manager's run, the proof of concept built with it
		MemoryMiB: 4096,
		// more than eight seldom speeds a build up and takes from the host
		CPUs:  min(runtime.NumCPU(), 8),
		Disks: disks,
		Card:  &b.Card,
	}
}
