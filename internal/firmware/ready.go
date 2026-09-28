package firmware

import (
	"context"

	"github.com/The127/miso/internal/download"
)

// Ready fetches miso's pinned OVMF package, through the store, and takes
// the firmware out of it.
func Ready(ctx context.Context, store *download.Store) (Firmware, error) {
	return ready(ctx, store, pin)
}

// ready takes the firmware out of the package a pin names.
func ready(ctx context.Context, store *download.Store, pin download.Pin) (Firmware, error) {
	pkg, err := store.OpenPinned(ctx, pin)
	if err != nil {
		return Firmware{}, err
	}

	defer func() { _ = pkg.Close() }()

	return read(pkg)
}
