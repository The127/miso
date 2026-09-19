package agent

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// enterRoot makes root the root of the run, with nothing of the builder
// above it.
func enterRoot(root string) error {
	if err := os.Chdir(root); err != nil {
		return err
	}

	// with both at ".", the old root lands on the new one and needs no
	// directory in the layers
	if err := unix.PivotRoot(".", "."); err != nil {
		return fmt.Errorf("pivot_root %s: %w", root, err)
	}

	// a chroot can be climbed out of, a root with nothing above it cannot
	if err := unix.Unmount(".", unix.MNT_DETACH); err != nil {
		return fmt.Errorf("detach the builder's root: %w", err)
	}

	return os.Chdir("/")
}
