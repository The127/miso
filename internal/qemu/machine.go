package qemu

import "io"

// Machine is a VM to start.
type Machine struct {
	Boot        Boot
	MemoryMiB   int
	CPUs        int
	Disks       []Disk
	Card        *Card
	Console     io.Writer
	Credentials []Credential

	// the machine is QEMU's microvm board, with virtio devices on a bus
	// that has no PCI, like a microVM of another VMM has
	Microvm bool

	// where QEMU keeps its temporary files, such as what a snapshot disk
	// took of a boot, the host's temporary directory if empty
	Temp string
}
