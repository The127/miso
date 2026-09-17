package imagefile

import (
	"errors"
	"strings"
)

func readFrom(line int, arguments string) (Stage, error) {
	fields := strings.Fields(arguments)
	named := len(fields) == 3 && strings.EqualFold(fields[1], "AS")

	switch {
	case len(fields) == 0:
		return Stage{}, errors.New("needs a base")
	case len(fields) == 1:
		return Stage{Line: line, Base: fields[0]}, nil
	case named:
		return Stage{Line: line, Base: fields[0], Name: fields[2]}, nil
	default:
		return Stage{}, errors.New("takes a base and an optional AS name")
	}
}
