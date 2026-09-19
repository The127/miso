package builderkernel

import (
	"io"
	"strconv"
	"strings"
)

// dataMember is the file of a Debian package that holds its files.
const dataMember = "data.tar.xz"

// data reads through a package to its data, which follows the header of
// the member that holds it.
func data(deb io.Reader) (io.Reader, error) {
	if _, err := io.CopyN(io.Discard, deb, 8); err != nil {
		return nil, err
	}

	for {
		header := make([]byte, 60)
		if _, err := io.ReadFull(deb, header); err != nil {
			return nil, err
		}

		name := strings.TrimRight(strings.TrimSpace(string(header[:16])), "/")
		size, err := strconv.ParseInt(strings.TrimSpace(string(header[48:58])), 10, 64)
		if err != nil {
			return nil, err
		}

		if name == dataMember {
			return io.LimitReader(deb, size), nil
		}

		// a member is padded to an even length
		if _, err := io.CopyN(io.Discard, deb, size+size%2); err != nil {
			return nil, err
		}
	}
}
