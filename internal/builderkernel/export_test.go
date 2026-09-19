package builderkernel

import "io"

// Data is the data of a package, which holds the kernel and its modules.
func Data(deb io.Reader) (io.Reader, error) {
	return data(deb)
}

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
