package builderkernel

import (
	"bytes"
	"debug/elf"
	"errors"
	"io"
	"strings"
)

// info is what a module says about itself.
type info struct {
	Name    string
	Depends []string
}

// modinfo reads what a module says about itself from its .modinfo section,
// which holds key=value entries, one after the other.
func modinfo(ko io.ReaderAt) (info, error) {
	module, err := elf.NewFile(ko)
	if err != nil {
		return info{}, err
	}

	section := module.Section(".modinfo")
	if section == nil {
		return info{}, errors.New("not a kernel module: no .modinfo section")
	}

	entries, err := section.Data()
	if err != nil {
		return info{}, err
	}

	var said info
	for entry := range bytes.SplitSeq(entries, []byte{0}) {
		key, value, found := strings.Cut(string(entry), "=")
		if !found {
			continue
		}

		switch key {
		case "name":
			said.Name = value
		case "depends":
			if value != "" {
				said.Depends = strings.Split(value, ",")
			}
		}
	}

	return said, nil
}
