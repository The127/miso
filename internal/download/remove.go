package download

import (
	"errors"
	"io/fs"
	"os"
)

// Remove takes the blob with a digest away. One that is not there is gone
// already.
func (s *Store) Remove(digest string) error {
	err := os.Remove(s.Path(digest))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return err
}
