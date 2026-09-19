package builderkernel

import (
	"bytes"
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
