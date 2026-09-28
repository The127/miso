package firmware

import (
	"fmt"
	"io"

	"github.com/The127/miso/internal/deb"
)

// the code and vars of one build, both for a 4 MiB flash
const (
	codeFile = "usr/share/OVMF/OVMF_CODE_4M.fd"
	varsFile = "usr/share/OVMF/OVMF_VARS_4M.fd"
)

// read takes the firmware out of a package.
func read(pkg io.Reader) (Firmware, error) {
	var found Firmware
	wanted := map[string]*[]byte{codeFile: &found.Code, varsFile: &found.Vars}

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
		return Firmware{}, fmt.Errorf("the package holds no usable %s", codeFile)
	}

	if len(found.Vars) == 0 {
		return Firmware{}, fmt.Errorf("the package holds no usable %s", varsFile)
	}

	return found, nil
}
