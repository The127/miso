package buildcontext

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"strconv"
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
	var files []string
	err := fs.WalkDir(d.root.FS(), path, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}

		file, err := d.file(name)
		files = append(files, file)

		return err
	})
	if err != nil {
		return "", err
	}

	return hashed(files), nil
}

func (d *Dir) file(name string) (string, error) {
	info, statErr := d.root.Lstat(name)
	content, readErr := d.root.ReadFile(name)
	if err := errors.Join(statErr, readErr); err != nil {
		return "", err
	}

	perm := strconv.FormatUint(uint64(info.Mode().Perm()), 8)
	sum := sha256.Sum256(content)

	return hashed([]string{name, perm, hex.EncodeToString(sum[:])}), nil
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
