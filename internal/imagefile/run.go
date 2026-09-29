package imagefile

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Run is a shell command, kept verbatim. An offline run has no network.
type Run struct {
	Line    int
	Offline bool
	Command string
}

func readRun(line int, arguments string) ([]Instruction, error) {
	// options come first, the command after them stays verbatim
	var flags []string
	command := arguments
	for strings.HasPrefix(command, "--") {
		end := strings.IndexAny(command, " \t")
		if end < 0 {
			end = len(command)
		}

		flags = append(flags, command[:end])
		command = strings.TrimLeft(command[end:], " \t")
	}

	_, options, err := splitOptions(flags)
	if err != nil {
		return nil, err
	}

	for _, name := range slices.Sorted(maps.Keys(options)) {
		value := options[name]
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
