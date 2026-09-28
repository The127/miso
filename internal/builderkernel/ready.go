package builderkernel

import (
	"context"

	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/initramfs"
)

// Kernel is what the builder VM boots: the kernel image and the modules its
// agent loads, in the order it loads them.
type Kernel struct {
	Release string
	Image   []byte
	Modules []initramfs.Module
}

// Ready fetches miso's pinned kernel package, through the store, and takes
// it apart into what the builder VM boots.
func Ready(ctx context.Context, store *download.Store) (Kernel, error) {
	return ready(ctx, store, pin, needs...)
}

// ready takes the package a pin names apart into what a VM boots.
func ready(ctx context.Context, store *download.Store, pin download.Pin, want ...string) (Kernel, error) {
	deb, err := store.OpenPinned(ctx, pin)
	if err != nil {
		return Kernel{}, err
	}

	defer func() { _ = deb.Close() }()

	held, err := read(deb)
	if err != nil {
		return Kernel{}, err
	}

	modules, err := unpack(held, want...)
	if err != nil {
		return Kernel{}, err
	}

	return Kernel{Release: held.Release, Image: held.Image, Modules: modules}, nil
}
