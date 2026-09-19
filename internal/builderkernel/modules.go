package builderkernel

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/ulikunitz/xz"
)

const (
	// moduleExt ends the name of a kernel module, its packing follows
	moduleExt = ".ko"
	// modulePacking is how a package holds its modules
	modulePacking = ".xz"
)

// plain is what a module file is called unpacked, and whether the file holds
// a module at all. A module carries its packing as the last extension.
func plain(file string) (string, bool) {
	name := file
	if path.Ext(name) != moduleExt {
		name = strings.TrimSuffix(name, path.Ext(name))
	}

	return name, strings.HasSuffix(name, moduleExt)
}

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

		file := path.Base(header.Name)

		name, isModule := plain(file)
		if !isModule {
			continue
		}

		if file != name+modulePacking {
			return nil, fmt.Errorf("the package holds %s, miso reads %s", file, name+modulePacking)
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
