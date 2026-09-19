package agent

import (
	"os"
	"path/filepath"
)

// makeFloor makes in scratch the layer a run lies on below every other, with
// the run's mount points in it.
func makeFloor(scratch string) (string, error) {
	floor := filepath.Join(scratch, "floor")
	for _, dir := range []string{floor, filepath.Join(floor, "proc"), filepath.Join(floor, "sys"), filepath.Join(floor, "dev")} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			return "", err
		}
	}

	return floor, nil
}
