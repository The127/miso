package qemu

// arguments boots the machine's kernel directly, with no boot loader and no
// disk to boot from.
func arguments(machine Machine) []string {
	return []string{"-kernel", machine.Kernel, "-initrd", machine.Initramfs, "-append", machine.CommandLine}
}
