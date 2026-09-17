package buildcontext

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

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
	var entries []string
	err := fs.WalkDir(d.root.FS(), path, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		found, err := d.entry(name, below(path, name), entry.IsDir())
		entries = append(entries, found)

		return err
	})
	if err != nil {
		return "", err
	}

	return hashed(entries), nil
}

func (d *Dir) entry(name string, below string, isDir bool) (string, error) {
	info, statErr := d.root.Lstat(name)

	var content string
	var readErr error
	if !isDir {
		content, readErr = d.content(name)
	}

	if err := errors.Join(statErr, readErr); err != nil {
		return "", err
	}

	perm := strconv.FormatUint(uint64(info.Mode().Perm()), 8)

	return hashed([]string{below, perm, content}), nil
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
