package imagefile

import "strings"

// Run is a shell command, kept verbatim. An offline run has no network.
type Run struct {
	Line    int
	Offline bool
	Command string
}

func (Run) instruction() {}

func readRun(line int, arguments string) ([]Instruction, error) {
	command, offline := strings.CutPrefix(arguments, "--network=none ")
	command, err := shellCommand(command)
	if err != nil {
		return nil, err
	}

	return []Instruction{Run{Line: line, Offline: offline, Command: command}}, nil
}
