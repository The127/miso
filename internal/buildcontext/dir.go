package buildcontext

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ErrOutsideContext is a source that is not below the build context.
var ErrOutsideContext = errors.New("outside the build context")

// ErrSpecialFile is something in the build context that is no file, no
// directory and no link, such as a pipe or a device. A copy cannot carry it.
var ErrSpecialFile = errors.New("special file")

// Dir is a build context, a directory on the host.
type Dir struct {
	// the root keeps every path below the directory, also one that tries
	// to leave it through .. or a symlink
	root *os.Root

	// what the root says to a path that leaves it. Go does not export that
	// error, so Open asks for one
	escapes error
}

// Open takes the directory a build file sits in as its build context.
func Open(dir string) (*Dir, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}

	_, escapes := root.Lstat("..")

	return &Dir{root: root, escapes: errors.Unwrap(escapes)}, nil
}

// Close lets go of the directory.
func (d *Dir) Close() error {
	return d.root.Close()
}

// Digest says what a COPY of this path would put into an image.
func (d *Dir) Digest(path string) (string, error) {
	entries, err := d.entries(path)
	if err != nil {
		return "", err
	}

	return hashed(entries), nil
}

func (d *Dir) entries(source string) ([]string, error) {
	if !filepath.IsLocal(source) {
		return nil, ErrOutsideContext
	}

	// the walk takes one spelling of a path only
	source = filepath.Clean(source)

	info, err := d.root.Lstat(source)
	if errors.Is(err, d.escapes) {
		return nil, ErrOutsideContext
	}

	if err != nil {
		return nil, err
	}

	// the walk below would follow a link it starts on
	if info.Mode()&fs.ModeSymlink != 0 {
		link, err := d.entry(source, ".")
		if err != nil {
			return nil, err
		}

		return []string{link}, nil
	}

	var entries []string
	err = fs.WalkDir(d.root.FS(), source, func(name string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		found, err := d.entry(name, below(source, name))
		if err != nil {
			return err
		}

		entries = append(entries, found)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// entry takes what a thing is and its mode from one look at it. The
// directory listing is older, and the thing may have been swapped since.
func (d *Dir) entry(name string, inside string) (string, error) {
	info, err := d.root.Lstat(name)
	if err != nil {
		return "", err
	}

	mode := info.Mode()
	switch {
	case mode.IsDir():
		return hashed([]string{"directory", inside, permissions(mode), ""}), nil
	case mode&fs.ModeSymlink != 0:
		// a link means a place in the image, not on the host, so it is
		// never followed here
		target, err := d.root.Readlink(name)
		if err != nil {
			return "", err
		}

		return hashed([]string{"link", inside, permissions(mode), target}), nil
	case mode.IsRegular():
		sum, err := d.content(name)
		if err != nil {
			return "", err
		}

		return hashed([]string{"file", inside, permissions(mode), sum}), nil
	default:
		// opening a pipe would wait for a writer forever
		return "", fmt.Errorf("%s: %w", name, ErrSpecialFile)
	}
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

func (d *Dir) content(name string) (string, error) {
	content, err := d.root.ReadFile(name)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(content)

	return hex.EncodeToString(sum[:]), nil
}

// below is a path as seen from the source of the copy. The source's own
// name stays out of a digest, the instruction already says it.
func below(source string, name string) string {
	if name == source {
		return "."
	}

	return strings.TrimPrefix(name, source+"/")
}

// hashed puts the length in front of every field, so that no field can
// run into the next.
func hashed(fields []string) string {
	hash := sha256.New()
	for _, field := range fields {
		hash.Write([]byte(strconv.Itoa(len(field)) + ":" + field))
	}

	return hex.EncodeToString(hash.Sum(nil))
}
