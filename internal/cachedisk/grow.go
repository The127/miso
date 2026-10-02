package cachedisk

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Grow makes the cache disk at a path larger, to a size in bytes, and keeps
// what is on it. Only under the Lock, as Make is, because a build writes to
// the disk. A disk only grows, a smaller one would cut the file system.
func Grow(path string, size int64) error {
	if err := grow(path, size); err != nil {
		return fmt.Errorf("grow cache disk %s: %w", path, err)
	}

	return nil
}

// grow is Grow without naming the disk in its errors.
func grow(path string, size int64) error {
	// before anything is touched, a missing tool leaves all as it was
	check, err := findTool("e2fsck")
	if err != nil {
		return err
	}

	resize, err := findTool("resize2fs")
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if size <= info.Size() {
		return fmt.Errorf("the disk is %d bytes and %d were asked for, a disk only grows", info.Size(), size)
	}

	// resize2fs only takes a file system that was checked, and a killed build
	// may have left it unclean. Preen fixes what is safe to fix
	if err := runTool(check, "-f", "-p", path); err != nil && !corrected(err) {
		return err
	}

	// the file is sparse, so it takes no more space until layers are written
	if err := os.Truncate(path, size); err != nil {
		return err
	}

	// a file larger than its file system would be refused as already grown
	// on another try, which never resized
	if err := runTool(resize, path); err != nil {
		return errors.Join(err, os.Truncate(path, info.Size()))
	}

	return nil
}

// corrected is e2fsck saying it fixed what it found, which it does with an
// exit code of 1. Any other bit is a file system it could not fix.
func corrected(err error) bool {
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		return false
	}

	return exit.ExitCode() == 1
}
