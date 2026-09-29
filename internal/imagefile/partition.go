package imagefile

import (
	"errors"
	"fmt"
	"strings"
)

// Partition is one partition of the disks of its stage, with the settings
// systemd-repart takes in a definition. Which settings exist is not the
// parser's business.
type Partition struct {
	Line     int
	Name     string
	Settings []Setting
}

// Setting is a line of a repart definition. A key may come more than once,
// as CopyFiles does.
type Setting struct {
	Key   string
	Value string
}

func (Partition) instruction() {}

func readPartition(line int, arguments string) ([]Instruction, error) {
	fields := strings.Fields(arguments)
	if len(fields) == 0 {
		return nil, errors.New("needs a name")
	}

	var settings []Setting
	for _, field := range fields[1:] {
		key, value, isSetting := strings.Cut(field, "=")
		if !isSetting || key == "" {
			return nil, fmt.Errorf("takes settings of the form Key=Value, not %q", field)
		}

		settings = append(settings, Setting{Key: key, Value: value})
	}

	return []Instruction{Partition{Line: line, Name: fields[0], Settings: settings}}, nil
}
