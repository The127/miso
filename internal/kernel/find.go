package kernel

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strconv"
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
	installed, err := installedIn(image)
	if err != nil {
		return Kernel{}, err
	}

	if len(installed) == 0 {
		return Kernel{}, fmt.Errorf("%w, looked at /%s/*/vmlinuz", ErrNoKernel, modules)
	}

	// the kernel systemd-boot would start first
	version := slices.MaxFunc(installed, vercmp.Compare)
	if wanted != "" {
		// a version only ever names a kernel the image has, never a path
		if !slices.Contains(installed, wanted) {
			return Kernel{}, fmt.Errorf("%w %s, the image has %s", ErrNoKernel, wanted, quoted(installed))
		}

		version = wanted
	}

	initrd, err := initrdOf(image, version)
	if err != nil {
		return Kernel{}, err
	}

	linux, _ := linuxOf(image, version)

	return Kernel{
		Version: version,
		Linux:   linux,
		Initrd:  initrd,
	}, nil
}

// quoted lists names from an image, which may come from anyone, with what
// could move a terminal's cursor or colour it escaped.
func quoted(names []string) string {
	each := make([]string, 0, len(names))
	for _, name := range names {
		each = append(each, strconv.Quote(name))
	}

	return strings.Join(each, ", ")
}

// installedIn are the versions of the kernels an image has. Modules a
// removed kernel left behind are none.
func installedIn(image fs.FS) ([]string, error) {
	versions, err := fs.ReadDir(image, modules)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	var installed []string
	for _, entry := range versions {
		if _, found := linuxOf(image, entry.Name()); found {
			installed = append(installed, entry.Name())
		}
	}

	return installed, nil
}
