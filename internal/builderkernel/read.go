package builderkernel

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
)

// kernelPrefix is what the kernel of a package is called, the release
// follows it.
const kernelPrefix = "vmlinuz-"

// contents is everything miso reads out of a kernel package.
type contents struct {
	Release string
	Image   []byte
	Builtin []string
	Modules []info
}

// read takes a package apart in one walk: the kernel it holds, what that
// kernel has built in, and what every module of it says about itself.
func read(deb io.Reader) (contents, error) {
	var held contents

	// an empty modules.builtin is not a missing one
	var listed bool

	files, failed := eachFile(deb)
	for name, file := range files {
		base := path.Base(name)

		switch {
		case path.Dir(name) == "boot" && strings.HasPrefix(base, kernelPrefix):
			release := strings.TrimPrefix(base, kernelPrefix)
			if held.Release != "" {
				return contents{}, fmt.Errorf("the package holds the kernels %s and %s", held.Release, release)
			}

			image, err := io.ReadAll(file)
			if err != nil {
				return contents{}, err
			}

			held.Release, held.Image = release, image

		case base == builtinFile:
			builtin, err := names(file)
			if err != nil {
				return contents{}, err
			}

			held.Builtin, listed = builtin, true

		default:
			ko, isModule := plain(base)
			if !isModule {
				continue
			}

			if base != ko+modulePacking {
				return contents{}, fmt.Errorf("the package holds %s, miso reads %s", base, ko+modulePacking)
			}

			// the bytes are kept packed, only the wanted modules are unpacked
			packed, err := io.ReadAll(file)
			if err != nil {
				return contents{}, err
			}

			said, err := says(bytes.NewReader(packed))
			if err != nil {
				return contents{}, err
			}

			said.Packed = packed
			held.Modules = append(held.Modules, said)
		}
	}

	if err := failed(); err != nil {
		return contents{}, err
	}

	if held.Release == "" {
		return contents{}, errors.New("the package holds no kernel")
	}

	if !listed {
		return contents{}, fmt.Errorf("the package holds no %s", builtinFile)
	}

	return held, nil
}
