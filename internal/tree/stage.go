package tree

import (
	"io/fs"
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/place"
)

// Land is where something at a path below a source lands in the image.
type Land func(source, below string, directory bool) (string, error)

// Into copies the sources of the stage whose root is at stage into the
// image, each where land says.
func Into(stage string, image *place.Root, sources []string, land Land) error {
	from, err := os.OpenRoot(stage)
	if err != nil {
		return err
	}

	defer func() { _ = from.Close() }()

	for _, source := range sources {
		if err := into(from, image, source, land); err != nil {
			return err
		}
	}

	return nil
}

func into(from *os.Root, image *place.Root, source string, land Land) error {
	in, err := from.Open(strings.TrimPrefix(source, "/"))
	if err != nil {
		return err
	}

	defer func() { _ = in.Close() }()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	target, err := land(source, ".", false)
	if err != nil {
		return err
	}

	if err := image.File(target, uint32(info.Mode().Perm()), in); err != nil {
		return err
	}

	kept, err := meta(source, info)
	if err != nil {
		return err
	}

	if kept.Xattrs, err = fileXattrs(in); err != nil {
		return &fs.PathError{Op: "getxattr", Path: source, Err: err}
	}

	return image.Keep(target, kept)
}

// meta is what a copy keeps of a file apart from its content.
func meta(source string, info fs.FileInfo) (place.Meta, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return place.Meta{}, &fs.PathError{Op: "stat", Path: source, Err: fs.ErrInvalid}
	}

	return place.Meta{
		UID:   stat.Uid,
		GID:   stat.Gid,
		Mode:  stat.Mode &^ unix.S_IFMT,
		Atime: unix.Timespec(stat.Atim),
		Mtime: unix.Timespec(stat.Mtim),
	}, nil
}
