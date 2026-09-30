package download

import (
	"os"
	"time"
)

// Use marks the blobs with some digests as used at a time.
func (s *Store) Use(at time.Time, digests ...string) error {
	for _, digest := range digests {
		if err := os.Chtimes(s.Path(digest), at, at); err != nil {
			return err
		}
	}

	return nil
}
