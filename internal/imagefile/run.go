package imagefile

import (
	"fmt"
	"strings"
)

// Run is a shell command, kept verbatim. An offline run has no network.
type Run struct {
	Line    int
	Offline bool
	Command string
}

func (Run) instruction() {}

func readRun(line int, arguments string) ([]Instruction, error) {
	// options come first, the command after them stays verbatim
	var flags []string
	command := arguments
	for strings.HasPrefix(command, "--") {
		var flag string
		flag, command, _ = strings.Cut(command, " ")
		flags = append(flags, flag)
	}

	_, options, err := splitOptions(flags)
	if err != nil {
		return nil, err
	}

	for name, value := range options {
		if name != "network" {
			return nil, fmt.Errorf("does not know --%s", name)
		}

		if value != "none" && value != "default" {
			return nil, fmt.Errorf("does not know --network=%s", value)
		}
	}

	command, err = shellCommand(command)
	if err != nil {
		return nil, err
	}

	return []Instruction{Run{Line: line, Offline: options["network"] == "none", Command: command}}, nil
}
