package tree

import (
	"io"
	"io/fs"
	"os"
	"path"
	"syscall"
)

// Copy copies what is in one directory into another.
func Copy(source, target string) error {
	// neither side may be left through a link
	from, err := os.OpenRoot(source)
	if err != nil {
		return err
	}

	defer func() { _ = from.Close() }()

	to, err := os.OpenRoot(target)
	if err != nil {
		return err
	}

	defer func() { _ = to.Close() }()

	c := copier{from: from, to: to, copied: map[inode]string{}}

	return c.entries(".")
}

// inode names a file apart from the names it has.
type inode struct {
	device, number uint64
}

type copier struct {
	from, to *os.Root
	// the name each file with several names was first copied as
	copied map[inode]string
}

func (c copier) entries(dir string) error {
	entries, err := fs.ReadDir(c.from.FS(), dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := path.Join(dir, entry.Name())

		info, err := entry.Info()
		if err != nil {
			return err
		}

		switch {
		case entry.IsDir():
			err = c.directory(name, info)
		case info.Mode()&fs.ModeSymlink != 0:
			err = c.link(name)
		default:
			err = c.fileOnce(name, info)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (c copier) directory(name string, info fs.FileInfo) error {
	if err := c.to.Mkdir(name, 0o700); err != nil {
		return err
	}

	if err := c.entries(name); err != nil {
		return err
	}

	// set once it is filled, which the mode may forbid and which moves the
	// time
	return c.keep(name, info)
}

// link copies a link as the text it holds and never follows it.
func (c copier) link(name string) error {
	text, err := c.from.Readlink(name)
	if err != nil {
		return err
	}

	return c.to.Symlink(text, name)
}

// fileOnce copies a file with several names once and links its other names
// to that copy.
func (c copier) fileOnce(name string, info fs.FileInfo) error {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Nlink < 2 {
		return c.file(name, info)
	}

	file := inode{device: stat.Dev, number: stat.Ino}
	if first, seen := c.copied[file]; seen {
		return c.to.Link(first, name)
	}

	c.copied[file] = name

	return c.file(name, info)
}

func (c copier) file(name string, info fs.FileInfo) error {
	in, err := c.from.Open(name)
	if err != nil {
		return err
	}

	defer func() { _ = in.Close() }()

	out, err := c.to.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()

		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	// set after the writing, which the mode may forbid and which moves the
	// time, and apart from the create, whose mode the umask changes
	return c.keep(name, info)
}

// keep gives a copy the mode and time of what it was copied from.
func (c copier) keep(name string, info fs.FileInfo) error {
	if err := c.to.Chmod(name, info.Mode().Perm()); err != nil {
		return err
	}

	return c.to.Chtimes(name, info.ModTime(), info.ModTime())
}
