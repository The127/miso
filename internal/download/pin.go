package download

import (
	"context"
	"os"
)

// Pin names bytes by their digest, and a URL they can be had from.
type Pin struct {
	URL    string
	Digest string
}

// OpenPinned opens the bytes a pin names, downloaded if the store does not
// hold them yet.
func (s *Store) OpenPinned(ctx context.Context, pin Pin) (*os.File, error) {
	path, err := s.Pinned(ctx, pin)
	if err != nil {
		return nil, err
	}

	return os.Open(path)
}
