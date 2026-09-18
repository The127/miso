package main

import (
	"os"
	"path/filepath"
)

// cacheDir is where miso keeps what it fetched, under the user's cache
// directory and never inside a build context.
func cacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "miso"), nil
}
