package qemu

// Disk is a disk image the machine sees as a virtio disk, which the agent
// finds by its serial.
type Disk struct {
	Path     string
	Format   string
	Serial   string
	ReadOnly bool
}
