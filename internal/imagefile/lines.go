package imagefile

import (
	"fmt"
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
			current.problem = fmt.Errorf("%w onto an empty line", ErrContinuation)
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
		current.problem = fmt.Errorf("%w runs off the end of the file", ErrContinuation)
		lines = append(lines, current)
	}

	return lines
}
