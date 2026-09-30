package download

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Swept is what a sweep removed.
type Swept struct {
	Count int
	Bytes int64
}

// Sweep removes the downloads that never arrived whole, which a crash
// leaves behind. Only while no download runs, or it takes one from under
// its writer.
func (s *Store) Sweep() (Swept, error) {
	var swept Swept

	names, err := os.ReadDir(s.blobs())
	if errors.Is(err, fs.ErrNotExist) {
		return swept, nil
	}

	if err != nil {
		return swept, err
	}

	for _, name := range names {
		if !strings.HasPrefix(name.Name(), temporary) {
			continue
		}

		info, err := name.Info()
		if err != nil {
			return swept, err
		}

		if err := os.Remove(filepath.Join(s.blobs(), name.Name())); err != nil {
			return swept, err
		}

		swept.Count++
		swept.Bytes += info.Size()
	}

	return swept, nil
}
