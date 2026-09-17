package imagefile

import (
	"errors"
	"fmt"
	"strings"
)

// Output turns the stage into a file of the given kind. Which kinds and
// options exist is not the parser's business.
type Output struct {
	Line    int
	Kind    string
	Name    string
	Options map[string]string
}

func (Output) instruction() {}

func readOutput(line int, arguments string) ([]Instruction, error) {
	var fields []string
	var options map[string]string
	for _, word := range strings.Fields(arguments) {
		option, isOption := strings.CutPrefix(word, "--")
		if !isOption {
			fields = append(fields, word)
			continue
		}

		if options == nil {
			options = map[string]string{}
		}

		name, value, _ := strings.Cut(option, "=")
		if _, twice := options[name]; twice {
			return nil, fmt.Errorf("has --%s twice", name)
		}

		options[name] = value
	}

	if len(fields) < 2 {
		return nil, errors.New("needs a kind and a file name")
	}

	if len(fields) > 2 {
		return nil, errors.New("takes a kind, a file name and options")
	}

	kind, name := fields[0], fields[1]
	return []Instruction{Output{Line: line, Kind: kind, Name: name, Options: options}}, nil
}
