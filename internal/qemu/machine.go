package qemu

import "io"

// Machine is what a VM boots: the builder's kernel, or through firmware
// an image miso made.
type Machine struct {
	Firmware    *Firmware
	Kernel      string
	Initramfs   string
	CommandLine string
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
