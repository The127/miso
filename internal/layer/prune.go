package layer

// Prune removes the layers the policy lets go. Only while nothing uses the
// store, like Sweep.
func (s *Store) Prune(policy Policy) (Swept, error) {
	var swept Swept

	entries, err := s.List()
	if err != nil {
		return swept, err
	}

	sizes := make(map[string]int64, len(entries))
	for _, entry := range entries {
		sizes[entry.Key] = entry.Size
	}

	for _, key := range Stale(entries, policy) {
		if err := s.Remove(key); err != nil {
			return swept, err
		}

		swept.Count++
		swept.Bytes += sizes[key]
	}

	return swept, nil
}
