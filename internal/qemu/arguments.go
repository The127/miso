package qemu

// arguments is how QEMU is told to run the machine.
func arguments(machine Machine) []string {
	return append(kernel(machine), size(machine)...)
}
