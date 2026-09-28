package kernel

import (
	"io/fs"
	"path"
)

// linuxes are where distributions put a kernel, most standard first.
// Debian 13 keeps it in /boot alone.
var linuxes = []func(version string) string{
	func(version string) string { return path.Join(modules, version, "vmlinuz") },
	func(version string) string { return "boot/vmlinuz-" + version },
}

// linuxOf is the first place of the kernel of a version that the image has.
func linuxOf(image fs.FS, version string) (string, bool) {
	for _, place := range linuxes {
		if _, err := fs.Stat(image, place(version)); err == nil {
			return place(version), true
		}
	}

	return "", false
}
