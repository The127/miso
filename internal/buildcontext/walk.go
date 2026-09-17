package buildcontext

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

type kind string

const (
	kindFile      kind = "file"
	kindDirectory kind = "directory"
	kindLink      kind = "link"
)

// entry is one thing a copy carries, and all that a digest or a payload
// may know of it.
type entry struct {
	kind kind

	// as seen from the source
	path string

	// empty for a link, which has none of its own that a copy could keep
	mode string

	// of a link, as written
	target string

	// where a file's content is, as seen from the build context
	name string
}

// walk visits the entries of a source in the one order every digest and
// every payload share.
func (d *Dir) walk(source string, visit func(entry) error) error {
	if !filepath.IsLocal(source) {
		return ErrOutsideContext
	}

	// the walk takes one spelling of a path only
	source = filepath.Clean(source)
	if err := d.throughLink(source); err != nil {
		return err
	}

	first, err := d.look(source, ".")
	if err != nil {
		return err
	}

	// the walk below would follow a link it starts on
	if first.kind == kindLink {
		return visit(first)
	}

	return fs.WalkDir(d.root.FS(), source, func(name string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		found, err := d.look(name, below(source, name))
		if err != nil {
			return err
		}

		return visit(found)
	})
}

// throughLink looks at every part of a source path but the last. The root
// would follow a link there as long as it stays inside.
func (d *Dir) throughLink(source string) error {
	parts := strings.Split(filepath.ToSlash(source), "/")
	for i := 1; i < len(parts); i++ {
		parent := strings.Join(parts[:i], "/")

		info, err := d.root.Lstat(parent)
		if err != nil {
			return err
		}

		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%s: %w", parent, ErrThroughLink)
		}
	}

	return nil
}

// look takes what a thing is and its mode from one look at it. A directory
// listing is older, and the thing may have been swapped since.
func (d *Dir) look(name string, path string) (entry, error) {
	info, err := d.root.Lstat(name)
	if err != nil {
		return entry{}, err
	}

	mode := info.Mode()
	switch {
	case mode.IsDir():
		return entry{kind: kindDirectory, path: path, mode: permissions(mode)}, nil
	case mode&fs.ModeSymlink != 0:
		// a link means a place in the image, not on the host, so it is
		// never followed here
		target, err := d.root.Readlink(name)
		if err != nil {
			return entry{}, err
		}

		return entry{kind: kindLink, path: path, target: target}, nil
	case mode.IsRegular():
		return entry{kind: kindFile, path: path, mode: permissions(mode), name: name}, nil
	default:
		// opening a pipe would wait for a writer forever
		return entry{}, fmt.Errorf("%s: %w", name, ErrSpecialFile)
	}
}

// below is a path as seen from the source of the copy. The source's own
// name stays out of a digest, the instruction already says it.
func below(source string, name string) string {
	if name == source {
		return "."
	}

	return strings.TrimPrefix(name, source+"/")
}

// permissions are the twelve bits of a unix mode in octal, like 4755. Go
// keeps the upper three apart from the nine that Perm gives.
func permissions(mode fs.FileMode) string {
	bits := uint64(mode.Perm())
	for flag, bit := range map[fs.FileMode]uint64{fs.ModeSetuid: 0o4000, fs.ModeSetgid: 0o2000, fs.ModeSticky: 0o1000} {
		if mode&flag != 0 {
			bits |= bit
		}
	}

	return strconv.FormatUint(bits, 8)
}
