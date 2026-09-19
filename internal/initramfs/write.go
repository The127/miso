package initramfs

import (
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
	"io"

	"github.com/u-root/u-root/pkg/cpio"
)

// Write writes an initial ramfs whose init is the program given, with the
// modules in the order to load them.
func Write(w io.Writer, init []byte, modules []Module) error {
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

	records := []cpio.Record{
		cpio.StaticRecord(init, cpio.Info{Name: "init", Mode: cpio.S_IFREG | 0o700}),
		// init has no output without it, and only some kernels bring one
		cpio.CharDev("dev/console", 0o600, 5, 1),
	}

	// the agent loads them in the order of their names
	for i, module := range modules {
		name := fmt.Sprintf("modules/%02d-%s.ko", i+1, module.Name)
		records = append(records, cpio.StaticRecord(module.Content, cpio.Info{Name: name, Mode: cpio.S_IFREG | 0o600}))
	}

	archive := cpio.Newc.Writer(w)
	if err := cpio.WriteRecordsAndDirs(archive, records); err != nil {
		return err
	}

	return cpio.WriteTrailer(archive)
}
