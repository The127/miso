package plan

import "github.com/The127/miso/internal/imagefile"

// Plan is a build file with everything looked up and keyed.
type Plan struct {
	Stages []Stage
}

// Stage is a stage of a plan.
type Stage struct {
	BaseDigest string
	Keys       []string
}

// New plans a build. Nothing is looked up again after it.
func New(stages []imagefile.Stage, agent string, context Context, bases Bases) (Plan, error) {
	if err := Validate(stages); err != nil {
		return Plan{}, err
	}

	p := &planner{agent: agent, context: context, bases: bases, ends: map[string]string{}}
	return p.plan(stages)
}
