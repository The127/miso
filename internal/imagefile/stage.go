package imagefile

// Stage is a FROM line and the instructions that follow it.
type Stage struct {
	Line         int
	Name         string
	Base         string
	Instructions []Instruction
}

// Instruction is one step of a stage: a Run, an Env or a Copy.
type Instruction interface {
	instruction()
}
