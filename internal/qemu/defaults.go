package qemu

// defaults leaves out every device QEMU adds on its own, and every config
// file of the host's QEMU, and names the board rather than taking the one
// the host's QEMU was built with, all of which differ from host to host, so
// the machine has only what it asks for.
func defaults(machine Machine) []string {
	board := "q35"
	if machine.Microvm {
		// the guest finds its virtio devices and powers off through ACPI, so
		// it is on whatever the host's QEMU defaults to
		board = "microvm,acpi=on"
	}

	return []string{"-nodefaults", "-no-user-config", "-machine", board}
}
