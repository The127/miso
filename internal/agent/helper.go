package agent

import (
	"fmt"
	"os"
	"syscall"
)

// helperName is the name the agent starts itself under to run a command.
const helperName = "miso-run"

// Helper sets up a run and becomes its shell when the agent was started as
// the helper of a run, and returns at once otherwise.
func Helper() {
	if os.Args[0] != helperName {
		return
	}

	root, command := os.Args[1], os.Args[2]
	if err := helper(root, command); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func helper(root, command string) error {
	if err := syscall.Chroot(root); err != nil {
		return fmt.Errorf("chroot %s: %w", root, err)
	}

	if err := os.Chdir("/"); err != nil {
		return err
	}

	// mounted from inside the run's PID namespace, so it shows the run's own
	if err := syscall.Mount("proc", "/proc", "proc", syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC, ""); err != nil {
		return fmt.Errorf("mount proc: %w", err)
	}

	// read-only, because the kernel behind it is the same for every later run
	if err := syscall.Mount("sysfs", "/sys", "sysfs", syscall.MS_RDONLY|syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC, ""); err != nil {
		return fmt.Errorf("mount sys: %w", err)
	}

	if err := mountDev(); err != nil {
		return err
	}

	return syscall.Exec("/bin/sh", []string{"/bin/sh", "-c", command}, os.Environ()) //nolint:gosec // running what the build file says is what a RUN is
}
