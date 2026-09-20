package builderkernel

import (
	"bytes"
	"io"

	"github.com/ulikunitz/xz"

	"github.com/The127/miso/internal/initramfs"
)

// unpack gives the wanted modules, with what they need, unpacked and in the
// order the agent loads them.
func unpack(held contents, want ...string) ([]initramfs.Module, error) {
	load, err := order(held.Modules, held.Builtin, want...)
	if err != nil {
		return nil, err
	}

	packed := make(map[string][]byte, len(held.Modules))
	for _, said := range held.Modules {
		packed[said.Name] = said.Packed
	}

	modules := make([]initramfs.Module, 0, len(load))
	for _, name := range load {
		file, err := xz.NewReader(bytes.NewReader(packed[name]))
		if err != nil {
			return nil, err
		}

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		modules = append(modules, initramfs.Module{Name: name, Content: content})
	}

	return modules, nil
}
