package firmware

import (
	"fmt"
	"io"

	"github.com/The127/miso/internal/deb"
)

// read takes the firmware out of the package of a source.
func read(pkg io.Reader, from source) (Firmware, error) {
	var found Firmware
	wanted := map[string]*[]byte{from.code: &found.Code, from.vars: &found.Vars}

	files, failed := deb.Files(pkg)
	for name, file := range files {
		into, isWanted := wanted[name]
		if !isWanted {
			continue
		}

		content, err := io.ReadAll(file)
		if err != nil {
			return Firmware{}, err
		}

		*into = content
	}

	if err := failed(); err != nil {
		return Firmware{}, err
	}

	// a link in a package reads as an empty file, and would boot nothing
	if len(found.Code) == 0 {
		return Firmware{}, fmt.Errorf("the package holds no usable %s", from.code)
	}

	if len(found.Vars) == 0 {
		return Firmware{}, fmt.Errorf("the package holds no usable %s", from.vars)
	}

	return found, nil
}
