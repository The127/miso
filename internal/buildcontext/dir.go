package buildcontext

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ErrOutsideContext is a source that is not below the build context.
var ErrOutsideContext = errors.New("outside the build context")

// Dir is a build context, a directory on the host.
type Dir struct {
	// the root keeps every path below the directory, also one that tries
	// to leave it through .. or a symlink
	root *os.Root
}

// Open takes the directory a build file sits in as its build context.
func Open(dir string) (*Dir, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}

	return &Dir{root: root}, nil
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

	info, err := d.root.Lstat(source)
	if err != nil {
		return nil, err
	}

	// the walk below would follow a link it starts on
	if info.Mode()&fs.ModeSymlink != 0 {
		link, err := d.entry(source, ".", fs.ModeSymlink)
		return []string{link}, err
	}

	var entries []string
	err = fs.WalkDir(d.root.FS(), source, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		found, err := d.entry(name, below(source, name), entry.Type())
		entries = append(entries, found)

		return err
	})

	return entries, err
}

func (d *Dir) entry(name string, below string, kind fs.FileMode) (string, error) {
	info, statErr := d.root.Lstat(name)

	called := "directory"

	var payload string
	var readErr error
	switch {
	case kind.IsDir():
	case kind&fs.ModeSymlink != 0:
		// a link means a place in the image, not on the host, so it is
		// never followed here
		called = "link"
		payload, readErr = d.root.Readlink(name)
	default:
		called = "file"
		payload, readErr = d.content(name)
	}

	if err := errors.Join(statErr, readErr); err != nil {
		return "", err
	}

	perm := strconv.FormatUint(uint64(info.Mode().Perm()), 8)

	return hashed([]string{called, below, perm, payload}), nil
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
