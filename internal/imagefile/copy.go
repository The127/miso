package imagefile

import "strings"

// Copy puts files from the build context into the image.
type Copy struct {
	Sources     []string
	Destination string
}

func (Copy) instruction() {}

func (p *parser) copy(arguments string) error {
	paths := strings.Fields(arguments)
	sources := paths[:len(paths)-1]
	destination := paths[len(paths)-1]
	return p.add("COPY", Copy{Sources: sources, Destination: destination})
}
