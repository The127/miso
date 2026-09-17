package imagefile

import (
	"errors"
	"fmt"
	"strings"
)

// Stage is a FROM line and the instructions that follow it.
type Stage struct {
	Name     string
	Base     string
	Commands []string
}

// Parse reads a build file into its stages, in order.
func Parse(source string) ([]Stage, error) {
	var stages []Stage
	for i, line := range strings.Split(source, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		keyword, rest, _ := strings.Cut(line, " ")
		switch keyword {
		case "FROM":
			base, name := from(line)
			stages = append(stages, Stage{Name: name, Base: base})
		case "RUN":
			if len(stages) == 0 {
				return nil, fmt.Errorf("line %d: RUN before FROM", i+1)
			}
			current := &stages[len(stages)-1]
			current.Commands = append(current.Commands, rest)
		case "":
		default:
			return nil, fmt.Errorf("line %d: unknown instruction %s", i+1, keyword)
		}
	}
	if len(stages) == 0 {
		return nil, errors.New("no FROM instruction")
	}
	return stages, nil
}

func from(line string) (base, name string) {
	fields := strings.Fields(line)
	base = fields[1]
	if len(fields) > 3 {
		name = fields[3]
	}
	return base, name
}
