package firmware

import (
	"context"
	"io"

	"github.com/The127/miso/internal/download"
)

// Read takes the firmware of an architecture out of a package.
func Read(pkg io.Reader, arch string) (Firmware, error) {
	return read(pkg, sources[arch])
}

// ReadyFrom takes the firmware of an architecture out of the package a pin
// names.
func ReadyFrom(ctx context.Context, store *download.Store, pin download.Pin, arch string) (Firmware, error) {
	from := sources[arch]
	from.pin = pin

	return ready(ctx, store, from)
}
