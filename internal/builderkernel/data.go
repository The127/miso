package builderkernel

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"iter"
	"path"
	"strings"

	"github.com/ulikunitz/xz"
)

// kernelPrefix is what the kernel of a package is called, the release
// follows it.
const kernelPrefix = "vmlinuz-"

// kernel is the release and the image of the kernel a package holds.
func kernel(deb io.Reader) (string, []byte, error) {
	var release string
	var image []byte

	files, failed := eachFile(deb)
	for name, file := range files {
		if path.Dir(name) != "boot" || !strings.HasPrefix(path.Base(name), kernelPrefix) {
			continue
		}

		found := strings.TrimPrefix(path.Base(name), kernelPrefix)
		if release != "" {
			return "", nil, fmt.Errorf("the package holds the kernels %s and %s", release, found)
		}

		release = found

		var err error
		if image, err = io.ReadAll(file); err != nil {
			return "", nil, err
		}
	}

	if err := failed(); err != nil {
		return "", nil, err
	}

	if release == "" {
		return "", nil, errors.New("the package holds no kernel")
	}

	return release, image, nil
}

// eachFile reads through the files of a package's data, giving the name of
// each and the file itself. A loop over it may stop early. Whatever went
// wrong is read from the second return, after the loop.
func eachFile(deb io.Reader) (iter.Seq2[string, io.Reader], func() error) {
	var failed error

	each := func(yield func(string, io.Reader) bool) {
		files, err := unpacked(deb)
		if err != nil {
			failed = err

			return
		}

		for {
			header, err := files.Next()
			if errors.Is(err, io.EOF) {
				return
			}

			if err != nil {
				failed = err

				return
			}

			// the names of a package start with ./, which Clean takes off
			if !yield(path.Clean(header.Name), files) {
				return
			}
		}
	}

	return each, func() error { return failed }
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
