package builderkernel

import (
	"archive/tar"
	"errors"
	"fmt"
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

	var release string
	var image []byte
	for {
		header, err := files.Next()
		if errors.Is(err, io.EOF) {
			break
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

		found := strings.TrimPrefix(name, kernelPrefix)
		if release != "" {
			return "", nil, fmt.Errorf("the package holds the kernels %s and %s", release, found)
		}

		release = found
		if image, err = io.ReadAll(files); err != nil {
			return "", nil, err
		}
	}

	if release == "" {
		return "", nil, errors.New("the package holds no kernel")
	}

	return release, image, nil
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
