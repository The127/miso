package imagefile

// Check is a shell command that must succeed inside the booted result,
// kept verbatim.
type Check struct {
	Line    int
	Command string
}

func (Check) instruction() {}

func readCheck(line int, arguments string) ([]Instruction, error) {
	command, err := shellCommand(arguments)
	if err != nil {
		return nil, err
	}

	return []Instruction{Check{Line: line, Command: command}}, nil
}
