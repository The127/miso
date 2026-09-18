package tree

import (
	"io"
	"io/fs"
	"os"
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

	entries, err := fs.ReadDir(from.FS(), ".")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return err
		}

		if err := copyFile(from, to, entry.Name(), info.Mode()); err != nil {
			return err
		}
	}

	return nil
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
