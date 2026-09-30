package download

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Pinned is where the bytes a pin names are, downloaded if the store does
// not hold them yet.
func (s *Store) Pinned(ctx context.Context, pin Pin) (string, error) {
	path := s.Path(pin.Digest)
	if _, err := os.Stat(path); err == nil {
		return path, s.Use(time.Now(), pin.Digest)
	}

	got, err := s.Get(ctx, pin.URL)
	if err != nil {
		return "", err
	}

	if got != pin.Digest {
		// the bytes would sit under their own digest, where a pin on that
		// digest would later find them without ever asking for them
		_ = os.Remove(s.Path(got))

		return "", fmt.Errorf("GET %s: pinned to %s, downloaded %s", pin.URL, pin.Digest, got)
	}

	return path, nil
}
