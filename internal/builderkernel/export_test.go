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
