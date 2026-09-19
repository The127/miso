package builderkernel

import "io"

// Data is the data of a package, which holds the kernel and its modules.
func Data(deb io.Reader) (io.Reader, error) {
	return data(deb)
}

// Kernel is the release and image of the kernel a package holds.
func Kernel(deb io.Reader) (string, []byte, error) {
	return kernel(deb)
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

// Modules is what every module a package holds says about itself.
func Modules(deb io.Reader) ([]Info, error) {
	return modules(deb)
}

// Builtin names everything the kernel of a package has built in.
func Builtin(deb io.Reader) ([]string, error) {
	return builtin(deb)
}
