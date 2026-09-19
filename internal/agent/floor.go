package agent

import (
	"os"
	"path/filepath"
)

// makeFloor makes in scratch the layer a run lies on below every other, with
// the run's mount points in it.
func makeFloor(scratch string) (string, error) {
	floor := filepath.Join(scratch, "floor")
	if err := os.Mkdir(floor, 0o700); err != nil {
		return "", err
	}

	for _, m := range mounts {
		if err := os.Mkdir(filepath.Join(floor, m.point), 0o700); err != nil {
			return "", err
		}
	}

	return floor, nil
}
