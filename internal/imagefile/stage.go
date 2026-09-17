package imagefile

// Stage is a FROM line and the instructions that follow it.
type Stage struct {
	Name         string
	Base         string
	Instructions []Instruction
}

// Instruction is one step of a stage: a Run or an Env.
type Instruction interface {
	instruction()
}

// Run is a shell command, kept verbatim.
type Run struct {
	Command string
}

// Env sets a variable for the instructions after it.
type Env struct {
	Key   string
	Value string
}

func (Run) instruction() {}
func (Env) instruction() {}
