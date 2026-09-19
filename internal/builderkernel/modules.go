package builderkernel

import (
	"bytes"
	"errors"
	"io"
	"path"
	"strings"

	"github.com/ulikunitz/xz"
)

// moduleSuffix is what a module of a package is called, its name comes first
const moduleSuffix = ".ko.xz"

// modules is what every module a package holds says about itself.
func modules(deb io.Reader) ([]info, error) {
	files, err := unpacked(deb)
	if err != nil {
		return nil, err
	}

	var found []info

	for {
		header, err := files.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		if !strings.HasSuffix(path.Base(header.Name), moduleSuffix) {
			continue
		}

		packed, err := xz.NewReader(files)
		if err != nil {
			return nil, err
		}

		ko, err := io.ReadAll(packed)
		if err != nil {
			return nil, err
		}

		said, err := modinfo(bytes.NewReader(ko))
		if err != nil {
			return nil, err
		}

		found = append(found, said)
	}

	return found, nil
}
