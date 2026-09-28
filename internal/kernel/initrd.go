package kernel

import (
	"io/fs"
	"path"
)

// initrds are where distributions put the initrd they make for a kernel,
// most standard first.
var initrds = []func(version string) string{
	func(version string) string { return path.Join(modules, version, "initrd") },
	func(version string) string { return "boot/initrd.img-" + version },
}

// initrdOf is the first place of an initrd for the kernel that the image has.
func initrdOf(image fs.FS, version string) string {
	for _, place := range initrds {
		if _, err := fs.Stat(image, place(version)); err == nil {
			return place(version)
		}
	}

	return ""
}
