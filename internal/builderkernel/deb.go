package builderkernel

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	// arMagic starts every ar archive, and a Debian package is one
	arMagic = "!<arch>\n"
	// dataMember is the file of a Debian package that holds its files
	dataMember = "data.tar.xz"
)

// data reads through a package to its data, which follows the header of
// the member that holds it.
func data(deb io.Reader) (io.Reader, error) {
	magic := make([]byte, len(arMagic))
	if _, err := io.ReadFull(deb, magic); err != nil || string(magic) != arMagic {
		return nil, errors.New("not a Debian package")
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

		if strings.HasPrefix(name, "data.tar.") {
			return nil, fmt.Errorf("the package holds %s, miso reads %s", name, dataMember)
		}

		// a member is padded to an even length
		if _, err := io.CopyN(io.Discard, deb, size+size%2); err != nil {
			return nil, err
		}
	}
}
