package imagefile

import (
	"errors"
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
	fields, options, err := splitOptions(strings.Fields(arguments))
	if err != nil {
		return nil, err
	}

	switch {
	case len(fields) < 2:
		return nil, errors.New("needs a kind and a file name")
	case len(fields) > 2:
		return nil, errors.New("takes a kind, a file name and options")
	}

	return []Instruction{Output{Line: line, Kind: fields[0], Name: fields[1], Options: options}}, nil
}
