package firmware

import (
	"context"
	"fmt"

	"github.com/The127/miso/internal/download"
)

// Ready fetches miso's pinned firmware package of the architecture, through
// the store, and takes the firmware out of it.
func Ready(ctx context.Context, store *download.Store, arch string) (Firmware, error) {
	from, ok := sources[arch]
	if !ok {
		return Firmware{}, fmt.Errorf("miso has no firmware for %s", arch)
	}

	return ready(ctx, store, from)
}

// ready takes the firmware out of the package of a source.
func ready(ctx context.Context, store *download.Store, from source) (Firmware, error) {
	pkg, err := store.OpenPinned(ctx, from.pin)
	if err != nil {
		return Firmware{}, err
	}

	defer func() { _ = pkg.Close() }()

	return read(pkg, from)
}
