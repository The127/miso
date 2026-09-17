package imagefile

import "strings"

// Output turns the stage into a file of the given kind. Which kinds exist
// is not the parser's business.
type Output struct {
	Line int
	Kind string
	Name string
}

func (Output) instruction() {}

func readOutput(line int, arguments string) ([]Instruction, error) {
	fields := strings.Fields(arguments)
	kind, name := fields[0], fields[1]
	return []Instruction{Output{Line: line, Kind: kind, Name: name}}, nil
}
