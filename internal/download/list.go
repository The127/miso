package download

import (
	"time"
)

// Blob is a file the store holds. Used is when it was last had or used.
type Blob struct {
	Digest string
	Used   time.Time
	Size   int64
}

// List gives the blobs the store holds. A download that has not arrived
// yet is no blob.
func (s *Store) List() ([]Blob, error) {
	names, err := s.entries()
	if err != nil {
		return nil, err
	}

	var blobs []Blob

	for _, name := range names {
		if unfinished(name) {
			continue
		}

		info, err := name.Info()
		if err != nil {
			return nil, err
		}

		blobs = append(blobs, Blob{Digest: "sha256:" + name.Name(), Used: info.ModTime(), Size: info.Size()})
	}

	return blobs, nil
}
