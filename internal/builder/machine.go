package builder

import (
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

// Machine is the builder VM a build boots, with its cache disk at a path.
func Machine(cache string) qemu.Machine {
	// the cache disk is an ext4 image as it is, which QEMU must never guess
	return qemu.Machine{
		Disks: []qemu.Disk{{Path: cache, Format: "raw", Serial: protocol.CacheSerial}},
	}
}
