package qemu

import "fmt"

// drives attaches each disk as a virtio disk, because only a virtio disk
// shows the agent its serial.
func drives(machine Machine) []string {
	var args []string
	for i, disk := range machine.Disks {
		id := fmt.Sprintf("disk%d", i)
		args = append(args,
			"-drive", fmt.Sprintf("file=%s,format=%s,if=none,id=%s,readonly=on", disk.Path, disk.Format, id),
			"-device", fmt.Sprintf("virtio-blk-pci,drive=%s,serial=%s", id, disk.Serial))
	}

	return args
}
