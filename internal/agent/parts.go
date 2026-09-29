package agent

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/kernel"
)

// part is a file of an image a UKI is built from, and the name the tools
// find its copy under.
type part struct {
	name     string
	path     string
	optional bool
}

// copyParts copies what a UKI of the kernel is built from out of the image
// into a directory, so the tools read the image's files and nothing that a
// link of the image would reach in their own root.
func copyParts(image fs.FS, found kernel.Kernel, dir string) error {
	parts := []part{
		{name: "linux", path: found.Linux},
		{name: "initrd", path: found.Initrd},
		{name: "stub", path: stub},
		{name: "os-release", path: "etc/os-release"},
		{name: "cmdline", path: "etc/kernel/cmdline", optional: true},
	}
	for _, p := range parts {
		err := copyPart(image, p.path, filepath.Join(dir, p.name))
		switch {
		case p.optional && errors.Is(err, fs.ErrNotExist):
			continue
		case p.name == "stub" && errors.Is(err, fs.ErrNotExist):
			return fmt.Errorf("the image has no UKI stub at /%s", stub)
		case err != nil:
			return err
		}
	}

	return nil
}

func copyPart(image fs.FS, path, to string) error {
	from, err := image.Open(path)
	if err != nil {
		return err
	}

	defer func() { _ = from.Close() }()

	into, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}

	_, err = io.Copy(into, from)

	return errors.Join(err, into.Close())
}
