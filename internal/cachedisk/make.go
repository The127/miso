package cachedisk

import "os/exec"

// Make makes a new cache disk at a path.
func Make(path string) error {
	return exec.Command("mkfs.ext4", "-q", path, "64M").Run() //nolint:gosec // the path is where miso keeps its own cache
}
