package qemu

import "slices"

// arguments is how QEMU is told to run the machine.
func arguments(machine Machine) []string {
	return slices.Concat(accel(), noReboot(), kernel(machine), size(machine), drives(machine))
}
