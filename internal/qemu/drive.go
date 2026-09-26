package qemu

import (
	"fmt"
	"strings"
)

// serialBytes is all of a serial a virtio disk shows, QEMU cuts a longer
// one without a word and the agent would look for a disk that is not there.
const serialBytes = 20

// drives attaches each disk as a virtio disk, because only a virtio disk
// shows the agent its serial.
func drives(machine Machine) ([]string, error) {
	var args []string
	for i, disk := range machine.Disks {
		if len(disk.Serial) > serialBytes {
			return nil, fmt.Errorf("the serial %s is longer than the %d bytes a virtio disk shows", disk.Serial, serialBytes)
		}

		id := fmt.Sprintf("disk%d", i)
		args = append(args,
			"-drive", fmt.Sprintf("file=%s,format=%s,if=none,id=%s,%s", escaped(disk.Path), disk.Format, id, access(disk)),
			"-device", fmt.Sprintf("virtio-blk-pci,drive=%s,serial=%s", id, escaped(disk.Serial)))
	}

	return args, nil
}

// access names the cache mode of a writable disk, because the cache disk
// holds a layer only once the guest's flushes reach it, which cache=unsafe
// would drop.
func access(disk Disk) string {
	if disk.ReadOnly {
		return "readonly=on"
	}

	return "cache=writeback"
}

// escaped keeps a comma in a value from starting the next option, QEMU
// reads a doubled comma as one.
func escaped(value string) string {
	return strings.ReplaceAll(value, ",", ",,")
}
