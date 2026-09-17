package imagefile

import (
	"errors"
	"strings"
)

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

		if err := p.add("ENV", Env{Key: key, Value: value}); err != nil {
			return err
		}
	}

	return nil
}
