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
