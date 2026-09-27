package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/baseimage"
	"github.com/The127/miso/internal/download"
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

// baseImages are the base images miso keeps in its cache, and the store
// their bytes live in.
func baseImages(cache string) (*download.Store, *baseimage.Cache) {
	blobs := download.Open(filepath.Join(cache, "bases"), http.DefaultClient)

	return blobs, baseimage.Open(filepath.Join(cache, "bases"), blobs, baseimage.Known)
}
