package qemu

// Machine is what a builder VM boots.
type Machine struct {
	Kernel      string
	Initramfs   string
	CommandLine string
}
