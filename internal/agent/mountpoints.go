package agent

import (
	"os"
	"path/filepath"
)

// mountPoints makes a layer of empty dev, proc and sys in scratch, to lie
// below every other layer of a rootfs. The floor of a run keeps them out of
// every layer, so an image from scratch has none.
func mountPoints(scratch string) (string, error) {
	points := filepath.Join(scratch, "mountpoints")
	if err := openDir(points); err != nil {
		return "", err
	}

	for _, name := range []string{"dev", "proc", "sys"} {
		if err := openDir(filepath.Join(points, name)); err != nil {
			return "", err
		}
	}

	return points, nil
}

// openDir makes a directory that is open to all, which a umask cannot narrow.
func openDir(path string) error {
	if err := os.Mkdir(path, 0o700); err != nil {
		return err
	}

	return os.Chmod(path, 0o755) //nolint:gosec // an image's mount points are open to all
}
