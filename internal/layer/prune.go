package layer

// Prune removes the layers the policy lets go. Only while nothing uses the
// store, like Sweep.
func (s *Store) Prune(policy Policy) error {
	entries, err := s.List()
	if err != nil {
		return err
	}

	for _, key := range Stale(entries, policy) {
		if err := s.Remove(key); err != nil {
			return err
		}
	}

	return nil
}
