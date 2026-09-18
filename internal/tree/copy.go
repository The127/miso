package tree

import (
	"io"
	"io/fs"
	"os"
	"path"
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

	return copyEntries(from, to, ".")
}

func copyEntries(from, to *os.Root, dir string) error {
	entries, err := fs.ReadDir(from.FS(), dir)
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
			err = copyDirectory(from, to, name, info.Mode())
		case info.Mode()&fs.ModeSymlink != 0:
			err = copyLink(from, to, name)
		default:
			err = copyFile(from, to, name, info.Mode())
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func copyDirectory(from, to *os.Root, name string, mode fs.FileMode) error {
	if err := to.Mkdir(name, 0o700); err != nil {
		return err
	}

	if err := copyEntries(from, to, name); err != nil {
		return err
	}

	// set once it is filled, which the mode may forbid
	return to.Chmod(name, mode.Perm())
}

// copyLink copies a link as the text it holds and never follows it.
func copyLink(from, to *os.Root, name string) error {
	text, err := from.Readlink(name)
	if err != nil {
		return err
	}

	return to.Symlink(text, name)
}

func copyFile(from, to *os.Root, name string, mode fs.FileMode) error {
	in, err := from.Open(name)
	if err != nil {
		return err
	}

	defer func() { _ = in.Close() }()

	out, err := to.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
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

	// set after the writing, which the mode may forbid, and apart from the
	// create, whose mode the umask changes
	return to.Chmod(name, mode.Perm())
}
