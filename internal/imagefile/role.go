package imagefile

// Role is what an instruction does to the stage it is in.
type Role int

const (
	// Builds makes the root file system of the stage.
	Builds Role = iota

	// ForDisks says something about the disks of the stage and leaves the
	// root file system as it is.
	ForDisks

	// Artifact makes a file of the build and leaves the root file system as
	// it is.
	Artifact
)

// RoleOf is the role of an instruction.
func RoleOf(instruction Instruction) Role {
	return instruction.role()
}

func (Run) role() Role       { return Builds }
func (Env) role() Role       { return Builds }
func (Copy) role() Role      { return Builds }
func (Partition) role() Role { return ForDisks }
func (Cmdline) role() Role   { return ForDisks }
func (Output) role() Role    { return Artifact }
func (Check) role() Role     { return Artifact }
