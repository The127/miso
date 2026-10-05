package qemu

import (
	"errors"
	"fmt"
	"strings"
)

// Disk is a disk image the machine sees as a virtio disk, which the agent
// finds by its serial, or as a CD.
type Disk struct {
	Path   string
	Format string
	Serial string
	Access Access

	// the machine sees the image as a CD in an optical drive, which it can
	// never write
	CD bool
}

// Access is what a boot may do to a disk.
type Access int

const (
	// Writable keeps what the boot writes.
	Writable Access = iota

	// ReadOnly refuses what the boot writes.
	ReadOnly

	// Snapshot takes what the boot writes and forgets it after, so the file
	// stays as it was.
	Snapshot
)

// serialBytes is all of a serial a virtio disk shows, QEMU cuts a longer
// one without a word and the agent would look for a disk that is not there.
const serialBytes = 20

// drives attaches each disk as a virtio disk, because only a virtio disk
// shows the agent its serial.
func drives(machine Machine) ([]string, error) {
	var args []string
	for i, disk := range machine.Disks {
		id := fmt.Sprintf("disk%d", i)
		if disk.CD && !boardOf(machine).bootsFirmware {
			return nil, errors.New("a microvm has no optical drive to put a CD in")
		}

		if disk.CD {
			drive := []string{"-device", "ide-cd,drive=" + id}
			if boardOf(machine).scsiCD {
				drive = []string{"-device", "virtio-scsi-pci,id=scsi", "-device", "scsi-cd,bus=scsi.0,drive=" + id}
			}

			args = append(args, "-drive", fmt.Sprintf("file=%s,format=%s,if=none,id=%s,media=cdrom,readonly=on", escaped(disk.Path), disk.Format, id))
			args = append(args, drive...)

			continue
		}

		if len(disk.Serial) > serialBytes {
			return nil, fmt.Errorf("the serial %s is longer than the %d bytes a virtio disk shows", disk.Serial, serialBytes)
		}

		args = append(args,
			"-drive", fmt.Sprintf("file=%s,format=%s,if=none,id=%s,%s", escaped(disk.Path), disk.Format, id, access(disk)),
			"-device", fmt.Sprintf("%s,drive=%s,serial=%s", boardOf(machine).device("virtio-blk"), id, escaped(disk.Serial)))
	}

	return args, nil
}

// access names the cache mode of a writable disk, because the cache disk
// holds a layer only once the guest's flushes reach it, which cache=unsafe
// would drop.
func access(disk Disk) string {
	switch disk.Access {
	case ReadOnly:
		return "readonly=on"
	case Snapshot:
		return "snapshot=on"
	case Writable:
	}

	return "cache=writeback"
}

// escaped keeps a comma in a value from starting the next option, QEMU
// reads a doubled comma as one.
func escaped(value string) string {
	return strings.ReplaceAll(value, ",", ",,")
}
