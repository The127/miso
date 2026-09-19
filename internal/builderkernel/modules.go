package builderkernel

import (
	"bytes"
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

// says is what the module in a packed file says about itself.
func says(file io.Reader) (info, error) {
	packed, err := xz.NewReader(file)
	if err != nil {
		return info{}, err
	}

	module, err := io.ReadAll(packed)
	if err != nil {
		return info{}, err
	}

	return modinfo(bytes.NewReader(module))
}

// modules is what every module a package holds says about itself.
func modules(deb io.Reader) ([]info, error) {
	var found []info

	files, failed := eachFile(deb)
	for name, file := range files {
		base := path.Base(name)

		ko, isModule := plain(base)
		if !isModule {
			continue
		}

		if base != ko+modulePacking {
			return nil, fmt.Errorf("the package holds %s, miso reads %s", base, ko+modulePacking)
		}

		said, err := says(file)
		if err != nil {
			return nil, err
		}

		found = append(found, said)
	}

	if err := failed(); err != nil {
		return nil, err
	}

	return found, nil
}
