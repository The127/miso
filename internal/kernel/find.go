package kernel

import (
	"io/fs"
	"path"
)

// modules holds a directory per installed kernel, named by its version.
const modules = "usr/lib/modules"

// Kernel is what boots an image, as paths in the image.
type Kernel struct {
	Version string
	Linux   string
	Initrd  string
}

// Find finds the kernel of an image and the initrd made for it.
func Find(image fs.FS) (Kernel, error) {
	versions, err := fs.ReadDir(image, modules)
	if err != nil {
		return Kernel{}, err
	}

	version := versions[0].Name()

	return Kernel{
		Version: version,
		Linux:   path.Join(modules, version, "vmlinuz"),
		Initrd:  path.Join(modules, version, "initrd"),
	}, nil
}
