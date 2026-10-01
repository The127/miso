package layer

import "github.com/The127/miso/internal/lru"

// Prune removes the layers the policy lets go. Only while nothing uses the
// store, like Sweep.
func (s *Store) Prune(policy lru.Policy) (lru.Swept, error) {
	entries, err := s.List()
	if err != nil {
		return lru.Swept{}, err
	}

	items := make([]lru.Item, len(entries))
	for i, entry := range entries {
		items[i] = lru.Item{ID: entry.Key, Used: entry.Used, Size: entry.Size}
	}

	return lru.Prune(items, policy, s.Remove)
}
