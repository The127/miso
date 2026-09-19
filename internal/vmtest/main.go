package vmtest

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

// Main runs the tests of a package as the init of a VM, then powers the VM
// off. The last line it prints tells the host how they went.
func Main(m *testing.M) {
	if os.Getpid() != 1 {
		fmt.Fprintln(os.Stderr, "these tests run as the init of a VM, see just test-vm")
		os.Exit(1)
	}

	code := 1
	if err := func() error {
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
	}(); err != nil {
		fmt.Println(err)
	} else if err := mountAll(); err != nil {
		fmt.Println(err)
	} else if err := loadModules(); err != nil {
		fmt.Println(err)
	} else {
		code = m.Run()
	}

	fmt.Printf("miso-vmtest: exit %d\n", code)
	syscall.Sync()

	// init must not return, the kernel panics when it does
	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF); err != nil {
		fmt.Println(err)
	}

	select {}
}

func loadModules() error {
	names, err := filepath.Glob("/modules/*.ko.xz")
	if err != nil {
		return err
	}

	for _, name := range names {
		if err := loadModule(name); err != nil {
			return fmt.Errorf("load %s: %w", name, err)
		}
	}

	return nil
}

func loadModule(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return unix.FinitModule(int(f.Fd()), "", unix.MODULE_INIT_COMPRESSED_FILE)
}

func mountAll() error {
	for _, m := range []struct{ fstype, target string }{
		{"proc", "/proc"},
		{"sysfs", "/sys"},
		{"devtmpfs", "/dev"},
		{"tmpfs", "/tmp"},
	} {
		if err := os.MkdirAll(m.target, 0o750); err != nil {
			return err
		}

		if err := syscall.Mount(m.fstype, m.target, m.fstype, 0, ""); err != nil {
			return fmt.Errorf("mount %s: %w", m.target, err)
		}
	}

	return nil
}
