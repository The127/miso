package builder

import (
	"fmt"
	"runtime"

	"github.com/The127/miso/internal/build"
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
	Requests []build.Request

	// where the base image with a digest is
	Blob func(digest string) string

	// the format the base image with a digest was fetched in
	Format func(digest string) (string, error)

	// the card its runs reach out through
	Card qemu.Card

	// the host's architecture, which the builder VM always runs. Empty is amd64.
	Arch string
}

// Machine is the builder VM of the build.
func (b Build) Machine() (qemu.Machine, error) {
	bases, err := b.bases()
	if err != nil {
		return qemu.Machine{}, err
	}

	// the cache disk is an ext4 image as it is, which QEMU must never guess
	cache := qemu.Disk{Path: b.Cache, Format: "raw", Serial: protocol.CacheSerial}

	return qemu.Machine{
		Boot: qemu.Kernel{
			Image:     b.Boot.Kernel,
			Initramfs: b.Boot.Initramfs,
			// the kernel hands init what follows --, and init is miso
			CommandLine: "console=ttyS0 panic=-1 -- agent",
		},
		// room for a package manager's run, the proof of concept built with it
		MemoryMiB: 4096,
		// more than eight seldom speeds a build up and takes from the host
		CPUs:  min(runtime.NumCPU(), 8),
		Disks: append([]qemu.Disk{cache}, bases...),
		Card:  &b.Card,
		Arch:  b.Arch,
	}, nil
}

// bases is a read-only disk for each base image the requests import.
func (b Build) bases() ([]qemu.Disk, error) {
	var disks []qemu.Disk
	for _, request := range b.Requests {
		request, isImport := request.Message.(protocol.Import)
		if !isImport {
			continue
		}

		format, err := b.Format(request.Digest)
		if err != nil {
			return nil, fmt.Errorf("base image %s: %w", request.Digest, err)
		}

		disks = append(disks, qemu.Disk{Path: b.Blob(request.Digest), Format: format, Serial: protocol.Serial(request.Digest), Access: qemu.ReadOnly})
	}

	return disks, nil
}
