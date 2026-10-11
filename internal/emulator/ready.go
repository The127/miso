package emulator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/The127/miso/internal/deb"
	"github.com/The127/miso/internal/download"
)

// programs are where a qemu-user package keeps the program that runs the
// programs of an architecture. Debian builds them static, so they run in
// the builder VM, which has no libraries.
var programs = map[string]string{
	"amd64": "usr/bin/qemu-x86_64",
	"arm64": "usr/bin/qemu-aarch64",
}

// Ready fetches miso's pinned qemu-user package of the builder's
// architecture, through the store, and takes the program out of it that
// runs the programs of target.
func Ready(ctx context.Context, store *download.Store, builder, target string) ([]byte, error) {
	pin, ok := pins[builder]
	if !ok {
		return nil, fmt.Errorf("miso has no QEMU for a builder on %s", builder)
	}

	if _, ok := programs[target]; !ok {
		return nil, fmt.Errorf("miso has no QEMU that runs programs of %s", target)
	}

	return ready(ctx, store, pin, target, emulators[builder][target])
}

// ready is the program that runs the programs of target, the one the store
// keeps under digest, or else taken out of the package a pin names.
func ready(ctx context.Context, store *download.Store, pin download.Pin, target, digest string) ([]byte, error) {
	if kept, err := os.ReadFile(store.Path(digest)); err == nil {
		return kept, store.Use(time.Now(), digest)
	}

	pkg, err := store.OpenPinned(ctx, pin)
	if err != nil {
		return nil, err
	}

	defer func() { _ = pkg.Close() }()

	wanted := programs[target]
	files, failed := deb.Files(pkg)
	for name, file := range files {
		if name == wanted {
			program, err := io.ReadAll(file)
			if err != nil {
				return nil, err
			}

			got, err := store.Keep(bytes.NewReader(program))
			if err != nil {
				return nil, err
			}

			if got != digest {
				// a pin on the digest it sits under would later find it
				_ = store.Remove(got)

				return nil, fmt.Errorf("the package holds %s as %s, pinned to %s", wanted, got, digest)
			}

			return program, nil
		}
	}

	if err := failed(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("the package holds no %s", wanted)
}
