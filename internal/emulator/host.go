package emulator

import (
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

// Host is the host's QEMU that runs the programs of target, found on PATH,
// and where it was found.
func Host(target string) (string, []byte, error) {
	name, known := statics[target]
	if !known {
		return "", nil, fmt.Errorf("miso has no QEMU that runs programs of %s", target)
	}

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		path := filepath.Join(dir, name)
		if found, err := os.ReadFile(path); err == nil { //nolint:gosec // the user's PATH is where their QEMU is
			return path, found, nil
		}
	}

	return "", nil, fmt.Errorf("a build for %s runs its steps through %s, which the qemu-user-static package installs", target, name)
}
