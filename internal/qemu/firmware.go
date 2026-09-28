package qemu

import "fmt"

// Firmware is a UEFI firmware the machine boots through, from its disks.
// Code stays as it is, Vars keeps what the firmware writes, so each boot
// needs a copy of its own.
type Firmware struct {
	Code string
	Vars string
}

func (f Firmware) arguments() []string {
	return []string{
		"-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", escaped(f.Code)),
		"-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", escaped(f.Vars)),
	}
}
