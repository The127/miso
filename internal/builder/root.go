package builder

import "os"

// inRoot runs a function on the directory as an os.Root, which keeps every
// name in it from leaving it.
func inRoot(dir string, do func(root *os.Root) error) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}

	defer func() { _ = root.Close() }()

	return do(root)
}
