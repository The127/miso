package imagefile

import (
	"errors"
	"fmt"
	"strings"
)

// Copy puts files into the image. They come from the build context, or
// from the stage named by From.
type Copy struct {
	From        string
	Sources     []string
	Destination string
}

func (Copy) instruction() {}

func (p *parser) copy(arguments string) error {
	var from string
	var paths []string
	for _, word := range strings.Fields(arguments) {
		flag, value, _ := strings.Cut(word, "=")
		switch {
		case !strings.HasPrefix(flag, "--"):
			paths = append(paths, word)
		case flag == "--from" && value == "":
			return errors.New("COPY --from needs a stage")
		case flag == "--from":
			from = value
		default:
			return fmt.Errorf("COPY does not know %s", flag)
		}
	}

	if len(paths) < 2 {
		return errors.New("COPY needs a source and a destination")
	}

	sources := paths[:len(paths)-1]
	destination := paths[len(paths)-1]
	return p.add("COPY", Copy{From: from, Sources: sources, Destination: destination})
}
