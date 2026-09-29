package imagefile

// Stage is a FROM line and the instructions that follow it.
type Stage struct {
	Line         int
	Name         string
	Base         string
	Instructions []Instruction
}

// Instruction is one step of a stage: a Run, an Env, a Copy, a Partition, a
// Cmdline, an Output or a Check. Each says what it does to its stage, and
// none can be added without.
type Instruction interface {
	role() Role
}
