package imagefile

import (
	"errors"
	"strings"
)

// Run is a shell command, kept verbatim.
type Run struct {
	Line    int
	Command string
}

func (Run) instruction() {}

func readRun(line int, arguments string) ([]Instruction, error) {
	if strings.TrimSpace(arguments) == "" {
		return nil, errors.New("needs a command")
	}

	return []Instruction{Run{Line: line, Command: arguments}}, nil
}
