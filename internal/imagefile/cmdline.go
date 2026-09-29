package imagefile

import (
	"errors"
	"strings"
)

// Cmdline is a part of the kernel command line of the disks of its stage,
// kept as it was written.
type Cmdline struct {
	Line int
	Text string
}

func readCmdline(line int, arguments string) ([]Instruction, error) {
	text := strings.TrimSpace(arguments)
	if text == "" {
		return nil, errors.New("needs some text")
	}

	return []Instruction{Cmdline{Line: line, Text: text}}, nil
}
