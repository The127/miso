package firmware

import (
	"context"
	"os"

	"github.com/The127/miso/internal/download"
)

// Ready fetches miso's pinned OVMF package, through the store, and takes
// the firmware out of it.
func Ready(ctx context.Context, store *download.Store) (Firmware, error) {
	return ready(ctx, store, pinURL, pinDigest)
}

// ready takes the firmware out of the package a pin names.
func ready(ctx context.Context, store *download.Store, url, digest string) (Firmware, error) {
	path, err := store.Pinned(ctx, url, digest)
	if err != nil {
		return Firmware{}, err
	}

	pkg, err := os.Open(path)
	if err != nil {
		return Firmware{}, err
	}

	defer func() { _ = pkg.Close() }()

	return read(pkg)
}
