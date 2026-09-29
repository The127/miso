package layer

import "os"

// Entry is a finished layer as the store lists it.
type Entry struct {
	Key string
}

// List gives the finished layers.
func (s *Store) List() ([]Entry, error) {
	names, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	var entries []Entry

	for _, name := range names {
		if unfinished(name.Name()) {
			continue
		}

		entries = append(entries, Entry{Key: name.Name()})
	}

	return entries, nil
}
