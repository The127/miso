package buildcontext

import (
	"os"
	"path/filepath"
)

// Dir is a build context, a directory on the host.
type Dir struct {
	dir string
}

// Open takes the directory a build file sits in as its build context.
func Open(dir string) *Dir {
	return &Dir{dir: dir}
}

// Digest says what a COPY of this path would put into an image.
func (d *Dir) Digest(path string) (string, error) {
	_, err := os.Lstat(filepath.Join(d.dir, path))
	return "", err
}
