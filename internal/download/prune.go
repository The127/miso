package download

import "github.com/The127/miso/internal/lru"

// Prune removes the blobs the policy lets go. Only while no download runs,
// like Sweep.
func (s *Store) Prune(policy lru.Policy) (Swept, error) {
	var swept Swept

	blobs, err := s.List()
	if err != nil {
		return swept, err
	}

	sizes := make(map[string]int64, len(blobs))
	items := make([]lru.Item, 0, len(blobs))

	for _, blob := range blobs {
		sizes[blob.Digest] = blob.Size
		items = append(items, lru.Item{ID: blob.Digest, Used: blob.Used, Size: blob.Size})
	}

	for _, digest := range lru.Stale(items, policy) {
		if err := s.Remove(digest); err != nil {
			return swept, err
		}

		swept.Count++
		swept.Bytes += sizes[digest]
	}

	return swept, nil
}
