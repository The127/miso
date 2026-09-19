package agent

import (
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/layer"
)

// Start mounts the cache disk with a serial on a directory and readies the
// layers on it for an agent.
func Start(serial, dir string) (*Agent, error) {
	if err := MountCache(serial, dir); err != nil {
		return nil, err
	}

	layers := filepath.Join(dir, "layers")
	if err := os.MkdirAll(layers, 0o700); err != nil {
		return nil, err
	}

	if err := layer.Open(layers).Sweep(); err != nil {
		return nil, err
	}

	// base disks are mounted on the VM's tmpfs, so a crash leaves no mount
	// points on the cache disk
	return New(layers, os.TempDir()), nil
}
