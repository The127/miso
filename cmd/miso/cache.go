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

	return blobs, baseimage.Open(filepath.Join(cache, "bases"), blobs, baseimage.Known["amd64"])
}

// cacheDiskSize is how large the cache disk may grow. The file is sparse,
// it takes only what the layers on it take. 20 GiB filled up under the KVM
// tests and a few builds of different images.
const cacheDiskSize = 50 << 30

// builderDir is where the builder keeps its cache, in the cache of miso.
func builderDir(cache string) cachedisk.Dir {
	return cachedisk.Dir(filepath.Join(cache, "builder"))
}

// holdCache holds the cache disk in a directory for this build, which makes
// the directory if it is not there yet.
func holdCache(dir cachedisk.Dir, said io.Writer) (io.Closer, error) {
	if err := os.MkdirAll(string(dir), 0o750); err != nil {
		return nil, err
	}

	return cachedisk.Lock(dir.LockFile(), func() {
		_, _ = fmt.Fprintln(said, "miso: waiting for another build to let go of the cache")
	})
}

// lockedCache holds the cache disk in a directory for this build, and makes
// the disk if it is not there yet.
func lockedCache(dir cachedisk.Dir, said io.Writer) (io.Closer, error) {
	held, err := holdCache(dir, said)
	if err != nil {
		return nil, err
	}

	disk := dir.Disk()
	if _, err := os.Stat(disk); errors.Is(err, fs.ErrNotExist) {
		err = cachedisk.Make(disk, cacheDiskSize)
		if err != nil {
			return nil, errors.Join(err, held.Close())
		}
	}

	return held, nil
}

// lockedDisk holds the cache disk of a directory and answers its file info as
// it is under the lock. It fails when there is no disk, and never makes one.
func lockedDisk(dir cachedisk.Dir, said io.Writer) (io.Closer, fs.FileInfo, error) {
	// the lock makes its file, so a missing disk is found before it
	if _, err := os.Stat(dir.Disk()); errors.Is(err, fs.ErrNotExist) {
		return nil, nil, fmt.Errorf("there is no cache disk at %s, the next build makes one", dir.Disk())
	}

	held, err := holdCache(dir, said)
	if err != nil {
		return nil, nil, err
	}

	// a build or another resize may have changed the disk while this one waited
	info, err := os.Stat(dir.Disk())
	if err != nil {
		return nil, nil, errors.Join(err, held.Close())
	}

	return held, info, nil
}
