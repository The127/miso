package qemu

import (
	"errors"
	"fmt"
	"strings"
)

// Boot is how a machine starts: a Kernel, or a Firmware that boots from the
// machine's disks. A machine never has both.
type Boot interface {
	arguments() []string
}

func boot(machine Machine) ([]string, error) {
	kernel, isKernel := machine.Boot.(Kernel)
	if !boardOf(machine).bootsFirmware && !isKernel {
		return nil, errors.New("a microvm needs a kernel to boot, it has no firmware to boot a disk")
	}

	if machine.Boot == nil {
		return nil, nil
	}

	// the credentials of a machine whose firmware brings no SMBIOS are on the
	// command line of its kernel
	if !boardOf(machine).firmware {
		words, err := commandLineCredentials(machine)
		if err != nil {
			return nil, err
		}

		kernel.CommandLine = strings.TrimSpace(kernel.CommandLine + " " + strings.Join(words, " "))
		if len(kernel.CommandLine) > commandLineMax {
			return nil, fmt.Errorf("the kernel command line is %d bytes with the credentials, the kernel keeps %d and would cut the rest", len(kernel.CommandLine), commandLineMax)
		}

		return kernel.arguments(), nil
	}

	return machine.Boot.arguments(), nil
}

// commandLineMax is how much of a command line an x86 kernel keeps. The
// kernel cuts the rest without a word, and a credential cut in the middle
// would be read as another one.
const commandLineMax = 2047
