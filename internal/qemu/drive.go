package qemu

import (
	"fmt"
	"strings"
)

// drives attaches each disk as a virtio disk, because only a virtio disk
// shows the agent its serial.
func drives(machine Machine) []string {
	var args []string
	for i, disk := range machine.Disks {
		id := fmt.Sprintf("disk%d", i)
		args = append(args,
			"-drive", fmt.Sprintf("file=%s,format=%s,if=none,id=%s,%s", escaped(disk.Path), disk.Format, id, access(disk)),
			"-device", fmt.Sprintf("virtio-blk-pci,drive=%s,serial=%s", id, disk.Serial))
	}

	return args
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
