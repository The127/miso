package qemu

import "strconv"

func size(machine Machine) []string {
	return []string{"-m", strconv.Itoa(machine.MemoryMiB) + "M", "-smp", strconv.Itoa(machine.CPUs)}
}
