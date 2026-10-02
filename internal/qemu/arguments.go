package qemu

import "slices"

// arguments is how QEMU is told to run the machine.
func arguments(machine Machine) ([]string, error) {
	disks, err := drives(machine)
	if err != nil {
		return nil, err
	}

	booting, err := boot(machine)
	if err != nil {
		return nil, err
	}

	return slices.Concat(defaults(machine), noReboot(), console(), booting, size(machine), disks, network(machine), credentials(machine)), nil
}
