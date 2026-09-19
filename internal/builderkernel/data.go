package builderkernel

import (
	"archive/tar"
	"errors"
	"io"
	"path"
	"strings"

	"github.com/ulikunitz/xz"
)

// kernelPrefix is what the kernel of a package is called, the release
// follows it.
const kernelPrefix = "vmlinuz-"

// kernel is the release and the image of the kernel a package holds.
func kernel(deb io.Reader) (string, []byte, error) {
	files, err := unpacked(deb)
	if err != nil {
		return "", nil, err
	}

	for {
		header, err := files.Next()
		if errors.Is(err, io.EOF) {
			return "", nil, errors.New("the package holds no kernel")
		}

		if err != nil {
			return "", nil, err
		}

		// the names of a package start with ./, which Clean takes off
		file := path.Clean(header.Name)
		name := path.Base(file)
		if path.Dir(file) != "boot" || !strings.HasPrefix(name, kernelPrefix) {
			continue
		}

		image, err := io.ReadAll(files)
		if err != nil {
			return "", nil, err
		}

		return strings.TrimPrefix(name, kernelPrefix), image, nil
	}
}

// unpacked reads the files of a package's data.
func unpacked(deb io.Reader) (*tar.Reader, error) {
	packed, err := data(deb)
	if err != nil {
		return nil, err
	}

	files, err := xz.NewReader(packed)
	if err != nil {
		return nil, err
	}

	return tar.NewReader(files), nil
}
