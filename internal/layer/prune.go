package layer

import "github.com/The127/miso/internal/lru"

// Prune removes the layers the policy lets go. Only while nothing uses the
// store, like Sweep.
func (s *Store) Prune(policy lru.Policy) (Swept, error) {
	var swept Swept

	entries, err := s.List()
	if err != nil {
		return swept, err
	}

	sizes := make(map[string]int64, len(entries))
	items := make([]lru.Item, 0, len(entries))

	for _, entry := range entries {
		sizes[entry.Key] = entry.Size
		items = append(items, lru.Item{ID: entry.Key, Used: entry.Used, Size: entry.Size})
	}

	for _, key := range lru.Stale(items, policy) {
		if err := s.Remove(key); err != nil {
			return swept, err
		}

		swept.Count++
		swept.Bytes += sizes[key]
	}

	return swept, nil
}
