package imagefile

// Run is a shell command, kept verbatim.
type Run struct {
	Line    int
	Command string
}

func (Run) instruction() {}

func readRun(line int, arguments string) ([]Instruction, error) {
	return []Instruction{Run{Line: line, Command: arguments}}, nil
}
