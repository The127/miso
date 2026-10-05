package qemu

import (
	"errors"
	"slices"
)

// arguments is how QEMU is told to run the machine.
func arguments(machine Machine) ([]string, error) {
	if machine.Arch == "arm64" && machine.Microvm {
		return nil, errors.New("an arm64 machine cannot be a microvm, QEMU has that board only for x86")
	}

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
