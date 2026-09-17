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

var errEnvUsage = errors.New("ENV needs KEY=VALUE")

func (p *parser) env(arguments string) error {
	assignments, closed := words(arguments)
	if !closed {
		return errors.New("ENV has an unclosed quote")
	}

	if len(assignments) == 0 {
		return errEnvUsage
	}

	for _, assignment := range assignments {
		key, value, assigned := strings.Cut(assignment, "=")
		if !assigned {
			return errEnvUsage
		}

		if err := p.add("ENV", Env{Line: p.line, Key: key, Value: value}); err != nil {
			return err
		}
	}

	return nil
}
