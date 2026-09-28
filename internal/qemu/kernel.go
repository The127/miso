package qemu

// Kernel is booted directly, with no boot loader and no disk to boot from.
type Kernel struct {
	Image       string
	Initramfs   string
	CommandLine string
}

func (k Kernel) arguments() []string {
	return []string{"-kernel", k.Image, "-initrd", k.Initramfs, "-append", k.CommandLine}
}
