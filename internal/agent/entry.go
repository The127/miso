package agent

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/The127/miso/internal/protocol"
)

// ErrUnknownSource is an entry of a source that the copy does not name.
var ErrUnknownSource = errors.New("of a source the copy does not name")

// ErrNotLocal is an entry whose path leaves its source.
var ErrNotLocal = errors.New("leaves its source")

// ErrOutOfOrder is an entry of a source after one of a later source. The
// sources of a copy come one after the other, in the order the build file
// names them.
var ErrOutOfOrder = errors.New("out of order")

// checked is an entry that belongs to a source the copy names, stays
// inside it, and comes after no entry of a later source than the last.
func checked(request protocol.Copy, entry protocol.Entry, last int) error {
	if entry.Source < 0 || entry.Source >= len(request.Sources) {
		return fmt.Errorf("%s: %w: %d", entry.Path, ErrUnknownSource, entry.Source)
	}

	if !filepath.IsLocal(entry.Path) {
		return fmt.Errorf("%s: %w", entry.Path, ErrNotLocal)
	}

	// each source is checked on its own, so its place among the others is
	// checked here, or the last to write would win
	if entry.Source < last {
		return fmt.Errorf("%s: %w: source %d after %d", entry.Path, ErrOutOfOrder, entry.Source, last)
	}

	return nil
}
