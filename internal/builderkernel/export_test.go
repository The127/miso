package builderkernel

import "io"

// Data is the data of a package, which holds the kernel and its modules.
func Data(deb io.Reader) (io.Reader, error) {
	return data(deb)
}
