package initramfs

import (
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
)

// machines maps an architecture to the ELF machine its builder VM runs.
var machines = map[string]elf.Machine{"amd64": elf.EM_X86_64, "arm64": elf.EM_AARCH64}

// errDynamic is a program that needs a dynamic loader.
var errDynamic = errors.New("needs a dynamic loader, which the builder VM does not have")

// checkInit refuses a program that cannot be the init of the builder VM.
func checkInit(arch string, init []byte) error {
	err := checkProgram(arch, "the init", init)
	if errors.Is(err, errDynamic) {
		return fmt.Errorf("%w: build miso with CGO_ENABLED=0", err)
	}

	return err
}

// checkProgram refuses a program, named by what it is, that the builder VM
// of an architecture cannot run.
func checkProgram(arch, name string, content []byte) error {
	machine, ok := machines[arch]
	if !ok {
		return fmt.Errorf("miso has no builder VM for %s", arch)
	}

	program, err := elf.NewFile(bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("%s is no program: %w", name, err)
	}

	if program.Machine != machine {
		return fmt.Errorf("%s is a program for %s, the builder VM runs %s", name, program.Machine, machine)
	}

	for _, segment := range program.Progs {
		if segment.Type == elf.PT_INTERP {
			return fmt.Errorf("%s %w", name, errDynamic)
		}
	}

	return nil
}
