package layer

import (
	"os"
	"time"
)

// Use marks the layers of some keys as used at a time.
func (s *Store) Use(at time.Time, keys ...string) error {
	for _, key := range keys {
		if err := checked(key); err != nil {
			return err
		}

		if err := os.Chtimes(s.Path(key), at, at); err != nil {
			return err
		}
	}

	return nil
}
