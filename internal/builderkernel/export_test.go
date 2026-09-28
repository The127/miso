package builderkernel

import (
	"context"
	"io"

	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/initramfs"
)

// Contents is everything miso reads out of a kernel package.
type Contents = contents

// Read takes a package apart in one walk.
func Read(deb io.Reader) (Contents, error) {
	return read(deb)
}

// Info is what a module says about itself.
type Info = info

// Modinfo is what a module says about itself.
func Modinfo(ko io.ReaderAt) (Info, error) {
	return modinfo(ko)
}

// Order is the order to load the wanted modules in.
func Order(have []Info, builtin []string, want ...string) ([]string, error) {
	return order(have, builtin, want...)
}

// Unpack gives the wanted modules unpacked and in the order they load.
func Unpack(held Contents, want ...string) ([]initramfs.Module, error) {
	return unpack(held, want...)
}

// ReadyFrom takes the package a pin names apart into what a VM boots.
func ReadyFrom(ctx context.Context, store *download.Store, url, digest string, want ...string) (Kernel, error) {
	return ready(ctx, store, url, digest, want...)
}

// Needs are the modules the builder VM loads.
var Needs = needs
