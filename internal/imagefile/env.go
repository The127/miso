package imagefile

import (
	"errors"
	"strings"
)

// Env sets a variable for the instructions after it.
type Env struct {
	Line  int
	Key   string
	Value string
}

func (Env) instruction() {}

var errEnvUsage = errors.New("needs KEY=VALUE")

func readEnv(line int, arguments string) ([]Instruction, error) {
	assignments, err := words(arguments)
	if err != nil {
		return nil, err
	}

	if len(assignments) == 0 {
		return nil, errEnvUsage
	}

	var variables []Instruction
	for _, assignment := range assignments {
		key, value, assigned := strings.Cut(assignment, "=")
		if !assigned {
			return nil, errEnvUsage
		}

		variables = append(variables, Env{Line: line, Key: key, Value: value})
	}

	return variables, nil
}
