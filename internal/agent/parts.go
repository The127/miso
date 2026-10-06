package agent

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/kernel"
	"github.com/The127/miso/internal/protocol"
)

// boot is what the tools build a UKI and a disk from: the kernel the image
// brings, the ESP made of its loader and the copies of its boot parts.
type boot struct {
	kernel kernel.Kernel
	esp    string
	parts  string
}

// prepareBoot makes the ESP and the copies of the image's boot parts under
// scratch. The command line of the request takes the place of the image's
// own, and a disk that boots from a CD gets a stub that says so.
func prepareBoot(imageFS fs.FS, image, scratch string, request protocol.Disk) (boot, error) {
	found, err := kernel.Find(imageFS, "")
	if err != nil {
		return boot{}, err
	}

	esp := filepath.Join(scratch, "esp")
	if err := makeESP(imageFS, image, esp); err != nil {
		return boot{}, err
	}

	parts := filepath.Join(scratch, "parts")
	if err := os.Mkdir(parts, 0o700); err != nil {
		return boot{}, err
	}

	if err := copyParts(imageFS, found, parts); err != nil {
		return boot{}, err
	}

	if request.Cmdline != "" {
		if err := os.WriteFile(filepath.Join(parts, "cmdline"), []byte(request.Cmdline+"\n"), 0o600); err != nil {
			return boot{}, err
		}
	}

	if request.ElTorito {
		if err := bootsFromCD(filepath.Join(parts, "stub")); err != nil {
			return boot{}, err
		}
	}

	return boot{kernel: found, esp: esp, parts: parts}, nil
}

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
		{name: "stub", path: imageEFI.Stub},
		{name: "os-release", path: "etc/os-release"},
		{name: "cmdline", path: "etc/kernel/cmdline", optional: true},
	}
	for _, p := range parts {
		err := copyPart(image, p.path, filepath.Join(dir, p.name))
		switch {
		case p.optional && errors.Is(err, fs.ErrNotExist):
			continue
		case p.name == "stub" && errors.Is(err, fs.ErrNotExist):
			return fmt.Errorf("the image has no UKI stub at /%s", imageEFI.Stub)
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

	info, err := from.Stat()
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		// a kernel's path holds a name from the image, which may come from
		// anyone
		return fmt.Errorf("the image's %q is no regular file", "/"+path)
	}

	into, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}

	_, err = io.Copy(into, from)

	return errors.Join(err, into.Close())
}
