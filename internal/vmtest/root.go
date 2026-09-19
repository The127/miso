package vmtest

import (
	"fmt"
	"os"
	"syscall"
)

// switchRoot moves the tests off the initial ramfs onto a tmpfs, because
// kernels before the empty mount under the initial ramfs refuse to
// pivot_root out of it, and a run pivots.
func switchRoot() error {
	if err := os.Mkdir("/vm", 0o700); err != nil {
		return err
	}

	if err := syscall.Mount("tmpfs", "/vm", "tmpfs", 0, ""); err != nil {
		return fmt.Errorf("mount /vm: %w", err)
	}

	init, err := os.ReadFile("/init")
	if err != nil {
		return err
	}

	if err := os.WriteFile("/vm/init", init, 0o700); err != nil { //nolint:gosec // init must be executable
		return err
	}

	if err := os.CopyFS("/vm/modules", os.DirFS("/modules")); err != nil {
		return err
	}

	if err := os.Chdir("/vm"); err != nil {
		return err
	}

	if err := syscall.Mount(".", "/", "", syscall.MS_MOVE, ""); err != nil {
		return fmt.Errorf("move /vm to /: %w", err)
	}

	if err := syscall.Chroot("."); err != nil {
		return err
	}

	return os.Chdir("/")
}
