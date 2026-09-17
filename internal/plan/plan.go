package plan

import "github.com/The127/miso/internal/imagefile"

// Plan is a build file with everything looked up and keyed.
type Plan struct {
	Stages []Stage
}

// Stage is a stage of a plan.
type Stage struct {
	BaseDigest string
	Steps      []Step
}

// Step is an instruction with the key of the layer it makes.
type Step struct {
	Instruction imagefile.Instruction
	Key         string
	Files       []File
}

// File is a file of the build context, as it was when the plan was made.
type File struct {
	Path   string
	Digest string
}

// New plans a build. Nothing is looked up again after it.
func New(stages []imagefile.Stage, agent string, context Context, bases Bases) (Plan, error) {
	if err := Validate(stages); err != nil {
		return Plan{}, err
	}

	p := &planner{agent: agent, context: context, bases: bases, ends: map[string]string{}}
	return p.plan(stages)
}
