package qemu

// Boot is how a machine starts: a Kernel, or a Firmware that boots from the
// machine's disks. A machine never has both.
type Boot interface {
	arguments() []string
}

func boot(machine Machine) []string {
	if machine.Boot == nil {
		return nil
	}

	return machine.Boot.arguments()
}
