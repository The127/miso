package imagefile

import (
	"errors"
	"fmt"
	"strings"
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

	var from string
	var paths []string
	for _, word := range found {
		flag, value, _ := strings.Cut(word, "=")
		switch {
		case !strings.HasPrefix(flag, "--"):
			paths = append(paths, word)
		case flag == "--from" && value == "":
			return nil, errors.New("--from needs a stage")
		case flag == "--from":
			from = value
		default:
			return nil, fmt.Errorf("does not know %s", flag)
		}
	}

	if len(paths) < 2 {
		return nil, errors.New("needs a source and a destination")
	}

	sources := paths[:len(paths)-1]
	destination := paths[len(paths)-1]
	return []Instruction{Copy{Line: line, From: from, Sources: sources, Destination: destination}}, nil
}
