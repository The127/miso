package builderkernel

import (
	"bytes"
	"debug/elf"
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

	entries, err := module.Section(".modinfo").Data()
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
