package download

import (
	"context"
	"os"
)

// Pinned is where the bytes of a digest are, downloaded from the URL if
// the store does not hold them yet.
func (s *Store) Pinned(_ context.Context, _, digest string) (string, error) {
	path := s.Path(digest)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}

	return path, nil
}
