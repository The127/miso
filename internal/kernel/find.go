package kernel

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"github.com/The127/miso/internal/vercmp"
)

// modules holds a directory per installed kernel, named by its version.
const modules = "usr/lib/modules"

// ErrNoKernel is an image without a kernel where distributions install it.
var ErrNoKernel = errors.New("no kernel")

// Kernel is what boots an image, as paths in the image.
type Kernel struct {
	Version string
	Linux   string
	Initrd  string
}

// Find finds the kernel of an image and the initrd made for it: the kernel
// of the wanted version, or the newest one when none is wanted.
func Find(image fs.FS, wanted string) (Kernel, error) {
	versions, err := fs.ReadDir(image, modules)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Kernel{}, err
	}

	var installed []string
	for _, entry := range versions {
		if _, err := fs.Stat(image, path.Join(modules, entry.Name(), "vmlinuz")); err == nil {
			installed = append(installed, entry.Name())
		}
	}

	if len(installed) == 0 {
		return Kernel{}, fmt.Errorf("%w, looked at /%s/*/vmlinuz", ErrNoKernel, modules)
	}

	// the kernel systemd-boot would start first
	version := slices.MaxFunc(installed, vercmp.Compare)
	if wanted != "" {
		// a version only ever names a kernel the image has, never a path
		if !slices.Contains(installed, wanted) {
			return Kernel{}, fmt.Errorf("%w %s, the image has %s", ErrNoKernel, wanted, strings.Join(installed, ", "))
		}

		version = wanted
	}

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
