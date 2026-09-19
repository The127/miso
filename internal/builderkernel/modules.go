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

		base := path.Base(header.Name)

		// a packed module carries its packing as the last extension
		plain := base
		if path.Ext(plain) != moduleExt {
			plain = strings.TrimSuffix(plain, path.Ext(plain))
		}

		if !strings.HasSuffix(plain, moduleExt) {
			continue
		}

		if base != plain+modulePacking {
			return nil, fmt.Errorf("the package holds %s, miso reads %s", base, plain+modulePacking)
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
