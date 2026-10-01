package download

import "github.com/The127/miso/internal/lru"

// Prune removes the blobs the policy lets go. Only while no download runs,
// like Sweep.
func (s *Store) Prune(policy lru.Policy) (lru.Swept, error) {
	blobs, err := s.List()
	if err != nil {
		return lru.Swept{}, err
	}

	items := make([]lru.Item, len(blobs))
	for i, blob := range blobs {
		items[i] = lru.Item{ID: blob.Digest, Used: blob.Used, Size: blob.Size}
	}

	return lru.Prune(items, policy, s.Remove)
}
