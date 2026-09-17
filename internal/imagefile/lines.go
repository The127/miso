package imagefile

import "strings"

// sourceLine is a line with its continuations joined in, numbered by where
// it starts in the file. It is unfinished when its last continuation has
// no line to continue on.
type sourceLine struct {
	number     int
	text       string
	unfinished bool
}

func sourceLines(source string) []sourceLine {
	var lines []sourceLine
	var current sourceLine
	for i, physical := range strings.Split(strings.TrimSuffix(source, "\n"), "\n") {
		if current.number == 0 {
			current.number = i + 1
		}

		head, continued := strings.CutSuffix(physical, "\\")
		current.text += head
		if continued {
			continue
		}

		lines = append(lines, current)
		current = sourceLine{}
	}

	if current.number != 0 {
		current.unfinished = true
		lines = append(lines, current)
	}

	return lines
}
