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

	var installed []string
	for _, entry := range versions {
		if _, err := fs.Stat(image, path.Join(modules, entry.Name(), "vmlinuz")); err == nil {
			installed = append(installed, entry.Name())
		}
	}

	version := installed[0]
	initrd, err := initrdOf(image, version)
	if err != nil {
		return Kernel{}, err
	}

	return Kernel{
		Version: version,
		Linux:   path.Join(modules, version, "vmlinuz"),
		Initrd:  initrd,
	}, nil
}
