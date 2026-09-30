package download

// Prune removes the blobs the policy lets go. Only while no download runs,
// like Sweep.
func (s *Store) Prune(policy Policy) (Swept, error) {
	var swept Swept

	blobs, err := s.List()
	if err != nil {
		return swept, err
	}

	sizes := make(map[string]int64, len(blobs))
	for _, blob := range blobs {
		sizes[blob.Digest] = blob.Size
	}

	for _, digest := range Stale(blobs, policy) {
		if err := s.Remove(digest); err != nil {
			return swept, err
		}

		swept.Count++
		swept.Bytes += sizes[digest]
	}

	return swept, nil
}
