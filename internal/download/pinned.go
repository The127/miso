package download

import (
	"context"
	"fmt"
	"os"
)

// Pinned is where the bytes of a digest are, downloaded from the URL if
// the store does not hold them yet.
func (s *Store) Pinned(ctx context.Context, url, digest string) (string, error) {
	path := s.Path(digest)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	got, err := s.Get(ctx, url)
	if err != nil {
		return "", err
	}

	if got != digest {
		// the bytes would sit under their own digest, where a pin on that
		// digest would later find them without ever asking for them
		_ = os.Remove(s.Path(got))

		return "", fmt.Errorf("GET %s: pinned to %s, downloaded %s", url, digest, got)
	}

	return path, nil
}
