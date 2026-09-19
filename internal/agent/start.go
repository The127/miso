package agent

import (
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
	if err := layer.Open(layers).Sweep(); err != nil {
		return nil, err
	}

	return nil, nil
}
