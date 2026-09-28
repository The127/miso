package layer

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ErrBadKey is a key that is no plain name, and would name something other
// than a layer of the store.
var ErrBadKey = errors.New("bad key")

// checked is a key that names a layer of the store and nothing else: one
// plain name, and none that the store's own directories are named with.
func checked(key string) error {
	plain := filepath.IsLocal(key) && key == filepath.Base(key) && key != "."
	if !plain || strings.HasPrefix(key, working) || strings.HasPrefix(key, scratch) {
		return fmt.Errorf("%q: %w", key, ErrBadKey)
	}

	return nil
}
