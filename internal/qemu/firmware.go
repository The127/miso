package qemu

import "fmt"

// Firmware is a UEFI firmware the machine boots through, from its disks,
// in place of a kernel. Code stays as it is, Vars keeps what the firmware
// writes, so each boot needs a copy of its own.
type Firmware struct {
	Code string
	Vars string
}

func boot(machine Machine) []string {
	if machine.Firmware == nil {
		return kernel(machine)
	}

	return []string{
		"-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", escaped(machine.Firmware.Code)),
		"-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", escaped(machine.Firmware.Vars)),
	}
}
