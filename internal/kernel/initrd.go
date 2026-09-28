package kernel

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// ErrNoInitrd is a kernel without an initrd in any place miso knows.
var ErrNoInitrd = errors.New("no initrd")

// initrds are where distributions put the initrd they make for a kernel,
// most standard first.
var initrds = []func(version string) string{
	func(version string) string { return path.Join(modules, version, "initrd") },
	func(version string) string { return "boot/initrd.img-" + version },
	func(version string) string { return "boot/initramfs-" + version + ".img" },
}

// initrdOf is the first place of an initrd for the kernel that the image has.
func initrdOf(image fs.FS, version string) (string, error) {
	looked := make([]string, 0, len(initrds))
	for _, place := range initrds {
		if _, err := fs.Stat(image, place(version)); err == nil {
			return place(version), nil
		}

		looked = append(looked, "/"+place(version))
	}

	return "", fmt.Errorf("%w for kernel %s, looked at %s", ErrNoInitrd, version, strings.Join(looked, ", "))
}
