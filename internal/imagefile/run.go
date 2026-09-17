package imagefile

// Run is a shell command, kept verbatim.
type Run struct {
	Line    int
	Command string
}

func (Run) instruction() {}

func readRun(line int, arguments string) ([]Instruction, error) {
	command, err := shellCommand(arguments)
	if err != nil {
		return nil, err
	}

	return []Instruction{Run{Line: line, Command: command}}, nil
}
