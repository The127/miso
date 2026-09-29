package layer

import (
	"os"
	"time"
)

// Entry is a finished layer as the store lists it. Used is when a build
// last made use of it.
type Entry struct {
	Key  string
	Used time.Time
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

		info, err := name.Info()
		if err != nil {
			return nil, err
		}

		entries = append(entries, Entry{Key: name.Name(), Used: info.ModTime()})
	}

	return entries, nil
}
