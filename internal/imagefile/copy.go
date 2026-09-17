package imagefile

import (
	"errors"
	"fmt"
)

// Copy puts files into the image. They come from the build context, or
// from the stage named by From.
type Copy struct {
	Line        int
	From        string
	Sources     []string
	Destination string
}

func (Copy) instruction() {}

func readCopy(line int, arguments string) ([]Instruction, error) {
	found, err := words(arguments)
	if err != nil {
		return nil, err
	}

	paths, options, err := splitOptions(found)
	if err != nil {
		return nil, err
	}

	for name := range options {
		if name != "from" {
			return nil, fmt.Errorf("does not know --%s", name)
		}
	}

	from, named := options["from"]
	if named && from == "" {
		return nil, errors.New("--from needs a stage")
	}

	if len(paths) < 2 {
		return nil, errors.New("needs a source and a destination")
	}

	sources := paths[:len(paths)-1]
	destination := paths[len(paths)-1]
	return []Instruction{Copy{Line: line, From: from, Sources: sources, Destination: destination}}, nil
}
