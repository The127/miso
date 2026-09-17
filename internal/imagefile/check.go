package imagefile

// Check is a shell command that must succeed inside the booted result,
// kept verbatim.
type Check struct {
	Line    int
	Command string
}

func (Check) instruction() {}

func readCheck(line int, arguments string) ([]Instruction, error) {
	return []Instruction{Check{Line: line, Command: arguments}}, nil
}
