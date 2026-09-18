package agent

import (
	"fmt"
	"syscall"
)

// mountSys gives a run a read-only /sys.
func mountSys() error {
	// read-only, because the kernel behind it is the same for every later run
	if err := syscall.Mount("sysfs", "/sys", "sysfs", syscall.MS_RDONLY|syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC, ""); err != nil {
		return fmt.Errorf("mount sys: %w", err)
	}

	return nil
}
