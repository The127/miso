package guestport

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Find is the device, under dev, of the VM's virtio port with the name, as
// the sysfs mounted at sys tells.
func Find(sys, dev, name string) (string, error) {
	ports := filepath.Join(sys, "class", "virtio-ports")

	devices, err := os.ReadDir(ports)
	if err != nil {
		return "", err
	}

	for _, device := range devices {
		named, err := os.ReadFile(filepath.Join(ports, device.Name(), "name"))
		if err != nil {
			return "", err
		}

		if strings.TrimSpace(string(named)) == name {
			return filepath.Join(dev, device.Name()), nil
		}
	}

	return "", fmt.Errorf("no virtio port named %q: %w", name, fs.ErrNotExist)
}
