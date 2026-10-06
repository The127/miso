package initramfs

import (
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
)

// machines maps an architecture to the ELF machine its builder VM runs.
var machines = map[string]elf.Machine{"amd64": elf.EM_X86_64, "arm64": elf.EM_AARCH64}

// checkInit refuses a program that cannot be the init of the builder VM.
func checkInit(arch string, init []byte) error {
	machine, ok := machines[arch]
	if !ok {
		return fmt.Errorf("miso has no builder VM for %s", arch)
	}

	program, err := elf.NewFile(bytes.NewReader(init))
	if err != nil {
		return fmt.Errorf("the init is no program: %w", err)
	}

	if program.Machine != machine {
		return fmt.Errorf("the init is a program for %s, the builder VM runs %s", program.Machine, machine)
	}

	for _, segment := range program.Progs {
		if segment.Type == elf.PT_INTERP {
			return errors.New("the init needs a dynamic loader, which the builder VM does not have: build miso with CGO_ENABLED=0")
		}
	}

	return nil
}
