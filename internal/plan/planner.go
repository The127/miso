package plan

import "github.com/The127/miso/internal/imagefile"

// planner is one planning run. Nothing else talks to the outside.
type planner struct {
	agent   string
	context Context
	bases   Bases
	ends    map[string]string
}

func (p *planner) plan(stages []imagefile.Stage) (Plan, error) {
	var planned Plan
	for _, stage := range stages {
		from, digest, err := p.start(stage)
		if err != nil {
			return Plan{}, err
		}

		steps, end, err := p.steps(from, stage)
		if err != nil {
			return Plan{}, err
		}

		planned.Stages = append(planned.Stages, Stage{BaseDigest: digest, Steps: steps})
		if stage.Name != "" {
			p.ends[stage.Name] = end
		}
	}

	return planned, nil
}

func (p *planner) start(stage imagefile.Stage) (key string, digest string, err error) {
	if end, onStage := p.ends[stage.Base]; onStage {
		return end, "", nil
	}

	digest, err = baseDigest(stage, p.bases)
	if err != nil {
		return "", "", err
	}

	return baseKey(p.agent, stage.Base, digest), digest, nil
}

// steps also hands back where the stage ends, which for a stage without
// steps is where it started.
func (p *planner) steps(from string, stage imagefile.Stage) ([]Step, string, error) {
	var steps []Step
	rootfs, end := from, from
	for _, instruction := range stage.Instructions {
		parent := rootfs
		switch instruction.(type) {
		case imagefile.Output, imagefile.Check:
			parent = end
		}

		step, err := p.step(parent, instruction)
		if err != nil {
			return nil, "", err
		}

		steps = append(steps, step)
		end = step.Key

		// an OUTPUT leaves the root file system as it is, so the steps
		// after it do not build on it
		if _, isOutput := instruction.(imagefile.Output); !isOutput {
			rootfs = step.Key
		}
	}

	return steps, end, nil
}

// step chains a step to its parent, so a change early in a stage reaches
// every key after it. Words and reads are hashed apart, so that neither
// list can run into the other.
func (p *planner) step(parent string, instruction imagefile.Instruction) (Step, error) {
	files, err := p.files(instruction)
	if err != nil {
		return Step{}, err
	}

	key := hashed([]string{parent, hashed(words(instruction)), hashed(p.reads(instruction, files))})
	return Step{Instruction: instruction, Key: key, Files: files}, nil
}

// files are what a step takes from the build context.
func (p *planner) files(instruction imagefile.Instruction) ([]File, error) {
	step, isCopy := instruction.(imagefile.Copy)
	if !isCopy || step.From != "" {
		return nil, nil
	}

	return contextFiles(step, p.context)
}

// reads are what a step takes from outside its own stage.
func (p *planner) reads(instruction imagefile.Instruction, files []File) []string {
	if step, isCopy := instruction.(imagefile.Copy); isCopy && step.From != "" {
		return []string{p.ends[step.From]}
	}

	var digests []string
	for _, file := range files {
		digests = append(digests, file.Digest)
	}

	return digests
}
