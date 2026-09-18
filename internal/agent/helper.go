package agent

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
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

	// its own, never the VM's, which shows the builder's disks and which a
	// run could break for every later run
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID, "mode=755"); err != nil {
		return fmt.Errorf("mount /dev: %w", err)
	}

	for _, device := range []struct {
		name         string
		major, minor uint32
	}{
		{"null", 1, 3},
		{"zero", 1, 5},
		{"full", 1, 7},
		{"random", 1, 8},
		{"urandom", 1, 9},
		{"tty", 5, 0},
	} {
		path := "/dev/" + device.name
		if err := unix.Mknod(path, unix.S_IFCHR, int(unix.Mkdev(device.major, device.minor))); err != nil { //nolint:gosec // the kernel keeps device numbers in 32 bits
			return fmt.Errorf("make %s: %w", path, err)
		}

		// after mknod, which the umask takes bits from
		if err := os.Chmod(path, 0o666); err != nil { //nolint:gosec // every user may use these devices, as on any Linux
			return err
		}
	}

	return syscall.Exec("/bin/sh", []string{"/bin/sh", "-c", command}, os.Environ()) //nolint:gosec // running what the build file says is what a RUN is
}
