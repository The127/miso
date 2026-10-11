package emulator

import (
	"bytes"
	"debug/elf"
	"fmt"
	"os"
	"path/filepath"
)

// statics are what qemu-user-static calls the program that runs the
// programs of an architecture, on Debian, Fedora and Arch alike.
var statics = map[string]string{
	"amd64": "qemu-x86_64-static",
	"arm64": "qemu-aarch64-static",
}

// Host is the host's QEMU that runs the programs of target, found on PATH.
func Host(target string) ([]byte, error) {
	name, known := statics[target]
	if !known {
		return nil, fmt.Errorf("miso has no QEMU that runs programs of %s", target)
	}

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		path := filepath.Join(dir, name)
		found, err := os.ReadFile(path) //nolint:gosec // the user's PATH is where their QEMU is
		if err != nil {
			continue
		}

		if err := staticRefusal(path, found); err != nil {
			return nil, err
		}

		return found, nil
	}

	return nil, fmt.Errorf("a build for %s runs its steps through %s, which the qemu-user-static package installs", target, name)
}

// staticRefusal is why a program cannot run in the builder VM, which has no
// libraries and no loader for them, or nil when it is built static.
func staticRefusal(path string, program []byte) error {
	file, err := elf.NewFile(bytes.NewReader(program))
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	for _, header := range file.Progs {
		if header.Type == elf.PT_INTERP {
			return fmt.Errorf("%s is not built static, so the builder VM cannot run it", path)
		}
	}

	return nil
}
