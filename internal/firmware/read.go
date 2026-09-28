package firmware

import (
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

	files, failed := deb.Files(pkg)
	for name, file := range files {
		switch name {
		case codeFile:
			code, err := io.ReadAll(file)
			if err != nil {
				return Firmware{}, err
			}

			found.Code = code
		case varsFile:
			vars, err := io.ReadAll(file)
			if err != nil {
				return Firmware{}, err
			}

			found.Vars = vars
		}
	}

	return found, failed()
}
