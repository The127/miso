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

	// where QEMU keeps its temporary files, such as what a snapshot disk
	// took of a boot, the host's temporary directory if empty
	Temp string
}
