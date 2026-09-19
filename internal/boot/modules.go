package boot

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func loadModules() error {
	// in the order of their numbers, which put each after what it needs
	names, err := filepath.Glob("/modules/*.ko")
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

	return unix.FinitModule(int(f.Fd()), "", 0)
}
