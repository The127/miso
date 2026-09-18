package agent

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// mountDev gives a run its own /dev with the usual devices and links.
func mountDev() error {
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

	for name, target := range map[string]string{
		"fd":     "/proc/self/fd",
		"stdin":  "/proc/self/fd/0",
		"stdout": "/proc/self/fd/1",
		"stderr": "/proc/self/fd/2",
	} {
		if err := os.Symlink(target, "/dev/"+name); err != nil {
			return err
		}
	}

	return nil
}
