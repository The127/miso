package tree

import (
	"io/fs"
	"os"
	"path"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/place"
)

// Land is where something at a path below a source lands in the image.
type Land func(source, below string, directory bool) (string, error)

// Into copies the sources of the stage whose root is at stage into the
// image, each where land says.
func Into(stage string, image *place.Root, sources []string, land Land) error {
	root, err := unix.Open(stage, unix.O_PATH|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return &fs.PathError{Op: "open", Path: stage, Err: err}
	}

	defer func() { _ = unix.Close(root) }()

	// one for all sources, so that a file named by two of them stays one
	copied := map[inode]string{}

	for _, source := range sources {
		if err := into(root, image, source, land, copied); err != nil {
			return err
		}
	}

	return nil
}

func into(stage int, image *place.Root, source string, land Land, copied map[inode]string) error {
	from, name, err := sourceRoot(stage, source)
	if err != nil {
		return err
	}

	defer func() { _ = from.Close() }()

	return stageCopy{from: from, image: image, source: source, land: land, copied: copied}.put(name, ".")
}

// stageCopy is one source of a stage on its way into the image.
type stageCopy struct {
	from   *os.Root
	image  *place.Root
	source string
	land   Land

	// where each file with several names first landed in the image
	copied map[inode]string
}

// put copies what is at a name of the stage, which is at a path below the
// source.
func (c stageCopy) put(name, below string) error {
	info, err := c.from.Lstat(name)
	if err != nil {
		return err
	}

	target, err := c.land(c.source, below, info.IsDir())
	if err != nil {
		return err
	}

	switch {
	case info.IsDir():
		return c.directory(name, below, target, info)
	case info.Mode()&fs.ModeSymlink != 0:
		return c.link(name, target, info)
	}

	return c.file(name, target, info)
}

// link copies a link as the text it holds and never follows it.
func (c stageCopy) link(name, target string, info fs.FileInfo) error {
	text, err := c.from.Readlink(name)
	if err != nil {
		return err
	}

	if err := c.image.Link(target, text); err != nil {
		return err
	}

	kept, err := meta(c.source, info)
	if err != nil {
		return err
	}

	// a link cannot be opened, its attributes are read through its directory
	err = at(c.from, "getxattr", name, func(path string) (err error) {
		kept.Xattrs, err = read(path)

		return err
	})
	if err != nil {
		return err
	}

	return c.image.Keep(target, kept)
}

func (c stageCopy) directory(name, below, target string, info fs.FileInfo) error {
	there, err := c.image.IsDirectory(target)
	if err != nil {
		return err
	}

	if err := c.image.Directory(target, uint32(info.Mode().Perm())); err != nil {
		return err
	}

	entries, err := fs.ReadDir(c.from.FS(), name)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := c.put(path.Join(name, entry.Name()), path.Join(below, entry.Name())); err != nil {
			return err
		}
	}

	// one the image has keeps its owner, mode and times, so that a stage
	// copied onto / does not hand them to /, /etc and the rest
	if there {
		return nil
	}

	opened, err := c.from.Open(name)
	if err != nil {
		return err
	}

	defer func() { _ = opened.Close() }()

	// once it is filled, since what lands in it moves its time
	return c.keep(target, info, opened)
}

func (c stageCopy) file(name, target string, info fs.FileInfo) error {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat.Nlink > 1 {
		file := inode{device: stat.Dev, number: stat.Ino}
		if first, seen := c.copied[file]; seen {
			return c.image.HardLink(target, first)
		}

		c.copied[file] = target
	}

	in, err := c.from.Open(name)
	if err != nil {
		return err
	}

	defer func() { _ = in.Close() }()

	if err := c.image.File(target, uint32(info.Mode().Perm()), in); err != nil {
		return err
	}

	return c.keep(target, info, in)
}

// keep gives what landed at the target what the opened original is apart
// from its content.
func (c stageCopy) keep(target string, info fs.FileInfo, opened *os.File) error {
	kept, err := meta(c.source, info)
	if err != nil {
		return err
	}

	if kept.Xattrs, err = fileXattrs(opened); err != nil {
		return &fs.PathError{Op: "getxattr", Path: c.source, Err: err}
	}

	return c.image.Keep(target, kept)
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
