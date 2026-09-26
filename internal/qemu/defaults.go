package qemu

// defaults leaves out every device QEMU adds on its own, so the machine has
// only what it asks for.
func defaults() []string {
	return []string{"-nodefaults"}
}
