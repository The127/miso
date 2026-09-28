package tree

import (
	"os"
	"strings"

	"github.com/The127/miso/internal/place"
)

// Land is where something at a path below a source lands in the image.
type Land func(source, below string, directory bool) (string, error)

// Into copies the sources of the stage whose root is at stage into the
// image, each where land says.
func Into(stage string, image *place.Root, sources []string, land Land) error {
	from, err := os.OpenRoot(stage)
	if err != nil {
		return err
	}

	defer func() { _ = from.Close() }()

	for _, source := range sources {
		if err := into(from, image, source, land); err != nil {
			return err
		}
	}

	return nil
}

func into(from *os.Root, image *place.Root, source string, land Land) error {
	in, err := from.Open(strings.TrimPrefix(source, "/"))
	if err != nil {
		return err
	}

	defer func() { _ = in.Close() }()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	target, err := land(source, ".", false)
	if err != nil {
		return err
	}

	return image.File(target, uint32(info.Mode().Perm()), in)
}
