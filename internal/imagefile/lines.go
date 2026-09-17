package imagefile

import (
	"errors"
	"strings"
)

// sourceLine is a line with its continuations joined in, numbered by where
// it starts in the file. It carries a problem when a continuation has no
// line to continue on.
type sourceLine struct {
	number  int
	text    string
	problem error
}

func sourceLines(source string) []sourceLine {
	var lines []sourceLine
	var current sourceLine
	continuing := false
	for i, physical := range strings.Split(strings.TrimSuffix(source, "\n"), "\n") {
		if !continuing {
			current = sourceLine{number: i + 1}
		}

		if continuing && strings.TrimSpace(physical) == "" {
			current.problem = errors.New("continuation onto an empty line")
			lines = append(lines, current)
			continuing = false
			continue
		}

		head, continued := strings.CutSuffix(physical, "\\")
		current.text += head
		continuing = continued
		if !continuing {
			lines = append(lines, current)
		}
	}

	if continuing {
		current.problem = errors.New("continuation runs off the end of the file")
		lines = append(lines, current)
	}

	return lines
}
