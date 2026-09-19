package boot

import (
	"fmt"
	"os"
	"syscall"
)

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
