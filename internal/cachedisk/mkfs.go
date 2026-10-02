package cachedisk

import "fmt"

// format has mkfs make an ext4 of a size in bytes at a path.
func format(mkfs, path string, size int64) error {
	// a bare number would count blocks
	kibibytes := fmt.Sprintf("%dk", size>>10)

	return runTool(mkfs, "-q", path, kibibytes)
}
