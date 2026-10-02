package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/baseimage"
	"github.com/The127/miso/internal/cachedisk"
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

// cacheDiskSize is how large the cache disk may grow. The file is sparse,
// it takes only what the layers on it take. 20 GiB filled up under the KVM
// tests and a few builds of different images.
const cacheDiskSize = 50 << 30

// lockedCache holds the cache disk in a directory for this build, and makes
// the disk if it is not there yet.
func lockedCache(dir string, said io.Writer) (io.Closer, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}

	held, err := cachedisk.Lock(filepath.Join(dir, "layers.lock"), func() {
		_, _ = fmt.Fprintln(said, "miso: waiting for another build to let go of the cache")
	})
	if err != nil {
		return nil, err
	}

	disk := filepath.Join(dir, "layers.img")
	if _, err := os.Stat(disk); errors.Is(err, fs.ErrNotExist) {
		err = cachedisk.Make(disk, cacheDiskSize)
		if err != nil {
			return nil, errors.Join(err, held.Close())
		}
	}

	return held, nil
}
