package sandbox

import (
	"fmt"
	"syscall"
)

// mountProc gives a run its own /proc, with the kernel settings read-only.
func mountProc() error {
	// mounted from inside the run's PID namespace, so it shows the run's own
	if err := syscall.Mount("proc", "/proc", "proc", syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC, ""); err != nil {
		return fmt.Errorf("mount proc: %w", err)
	}

	// the settings under it are the VM's kernel, the same for every later run
	if err := syscall.Mount("/proc/sys", "/proc/sys", "", syscall.MS_BIND, ""); err != nil {
		return fmt.Errorf("bind /proc/sys: %w", err)
	}

	if err := syscall.Mount("", "/proc/sys", "", syscall.MS_BIND|syscall.MS_REMOUNT|syscall.MS_RDONLY|syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC, ""); err != nil {
		return fmt.Errorf("make /proc/sys read-only: %w", err)
	}

	return nil
}
