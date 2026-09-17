package imagefile

import (
	"errors"
	"strings"
)

// Copy puts files from the build context into the image.
type Copy struct {
	Sources     []string
	Destination string
}

func (Copy) instruction() {}

func (p *parser) copy(arguments string) error {
	paths := strings.Fields(arguments)
	if len(paths) < 2 {
		return errors.New("COPY needs a source and a destination")
	}

	sources := paths[:len(paths)-1]
	destination := paths[len(paths)-1]
	return p.add("COPY", Copy{Sources: sources, Destination: destination})
}
