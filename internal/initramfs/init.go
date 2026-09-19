package initramfs

import (
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
)

// checkInit refuses a program that cannot be the init of the builder VM.
func checkInit(init []byte) error {
	program, err := elf.NewFile(bytes.NewReader(init))
	if err != nil {
		return fmt.Errorf("the init is no program: %w", err)
	}

	if program.Machine != elf.EM_X86_64 {
		return fmt.Errorf("the init is a program for %s, the builder VM runs %s", program.Machine, elf.EM_X86_64)
	}

	for _, segment := range program.Progs {
		if segment.Type == elf.PT_INTERP {
			return errors.New("the init needs a dynamic loader, which the builder VM does not have: build miso with CGO_ENABLED=0")
		}
	}

	return nil
}
