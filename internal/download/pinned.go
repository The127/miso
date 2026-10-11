package download

import (
	"context"
	"fmt"
	"io"
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

	if err := s.pinnedTo(pin.Digest, got); err != nil {
		return "", fmt.Errorf("GET %s: %w", pin.URL, err)
	}

	return path, nil
}

// KeepPinned keeps the bytes of body, which are pinned to digest.
func (s *Store) KeepPinned(digest string, body io.Reader) error {
	got, err := s.keep(body)
	if err != nil {
		return err
	}

	return s.pinnedTo(digest, got)
}

// pinnedTo refuses bytes kept under got that are pinned to digest. They
// would sit under their own digest, where a pin on that digest would later
// find them without ever asking for them.
func (s *Store) pinnedTo(digest, got string) error {
	if got == digest {
		return nil
	}

	_ = os.Remove(s.Path(got))

	return fmt.Errorf("pinned to %s, got %s", digest, got)
}
