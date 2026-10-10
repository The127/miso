package boot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// magics are what the header of a program of an architecture looks like to
// binfmt_misc, and the mask it is compared under, as Debian's qemu-user
// registers them.
var magics = map[string]struct{ magic, mask string }{
	"amd64": {
		magic: `\x7f\x45\x4c\x46\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x3e\x00`,
		mask:  `\xff\xff\xff\xff\xff\xfe\xfe\xfc\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff`,
	},
	"arm64": {
		magic: `\x7f\x45\x4c\x46\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\xb7\x00`,
		mask:  `\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff`,
	},
}

// registerEmulators registers the emulators init was handed. Only a build
// for another architecture has one, and only it needs binfmt_misc.
func registerEmulators() error {
	found, err := os.ReadDir("/emulators")
	if err != nil || len(found) == 0 {
		return err
	}

	if err := syscall.Mount("binfmt_misc", binfmtMisc, "binfmt_misc", 0, ""); err != nil {
		return fmt.Errorf("mount %s: %w", binfmtMisc, err)
	}

	return register("/emulators", writeEntry)
}

// writeEntry hands binfmt_misc an entry in one write, as it takes them, on
// a file it never truncates.
func writeEntry(entry string) error {
	f, err := os.OpenFile(filepath.Join(binfmtMisc, "register"), os.O_WRONLY, 0)
	if err != nil {
		return err
	}

	_, err = f.WriteString(entry)

	return errors.Join(err, f.Close())
}

// binfmtMisc is where the kernel takes binfmt_misc entries.
const binfmtMisc = "/proc/sys/fs/binfmt_misc"

// register has each emulator in dir, named by its architecture, run the
// programs of that architecture, through a write of its binfmt_misc entry.
// Debian's QEMU keeps a program's argv[0] only when its own name ends in
// -binfmt-P, so it is registered through a link of that name. F opens it
// right away, so it runs in the root of a run too.
func register(dir string, write func(entry string) error) error {
	found, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, emulator := range found {
		arch := emulator.Name()
		magic, known := magics[arch]
		if !known {
			return fmt.Errorf("no binfmt_misc entry for an emulator of %s", arch)
		}

		link := filepath.Join(dir, arch+"-binfmt-P")
		if err := os.Symlink(arch, link); err != nil {
			return err
		}

		if err := write(fmt.Sprintf(":%s:M::%s:%s:%s:OPF", arch, magic.magic, magic.mask, link)); err != nil {
			return err
		}
	}

	return nil
}
