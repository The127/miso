package layer

import "os"

// Remove takes the layer of a key away.
func (s *Store) Remove(key string) error {
	left, err := s.retire(key)
	if err != nil {
		return err
	}

	return os.RemoveAll(left)
}

// retire moves the layer out from under its key first, so a crash while its
// files go leaves only what Sweep clears and never a key with half a layer
func (s *Store) retire(key string) (string, error) {
	if err := checked(key); err != nil {
		return "", err
	}

	// Go refuses to rename over a directory, even an empty one
	left, err := s.Scratch()
	if err != nil {
		return "", err
	}

	if err := os.Remove(left); err != nil {
		return "", err
	}

	if err := os.Rename(s.Path(key), left); err != nil {
		return "", err
	}

	// the move must be on disk before the first file goes
	return left, syncFileSystem(left)
}
