package place

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// File puts a file with a content at a path of the image.
func (r *Root) File(path string, mode uint32, content io.Reader) error {
	file, err := os.OpenFile(filepath.Join(r.dir, path), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.FileMode(mode))
	if err != nil {
		return err
	}

	if _, err := io.Copy(file, content); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}
