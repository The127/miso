package qemu

// defaults leaves out every device QEMU adds on its own, and every config
// file of the host's QEMU, which differ from host to host, so the machine
// has only what it asks for.
func defaults() []string {
	return []string{"-nodefaults", "-no-user-config"}
}
