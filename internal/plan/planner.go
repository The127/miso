package plan

import (
	"slices"

	"github.com/The127/miso/internal/imagefile"
)

// planner is one planning run. Nothing else talks to the outside.
type planner struct {
	agent   string
	context Context
	bases   Bases
	ends    map[string]string
	outputs map[string]map[string]string
}

func (p *planner) plan(stages []imagefile.Stage) (Plan, error) {
	planned := Plan{Agent: p.agent}
	for _, stage := range stages {
		from, digest, err := p.start(stage)
		if err != nil {
			return Plan{}, err
		}

		steps, end, err := p.steps(from, stage)
		if err != nil {
			return Plan{}, err
		}

		if p.download(stage, digest) {
			planned.Downloads = append(planned.Downloads, stage.Base)
		}

		planned.Stages = append(planned.Stages, Stage{Name: stage.Name, Base: stage.Base, BaseDigest: digest, Steps: steps})
		if stage.Name != "" {
			p.ends[stage.Name] = end
			p.outputs[stage.Name] = outputKeys(steps)
		}
	}

	return planned, nil
}

// download is an image base that nobody has fetched yet. A stage has no
// digest either, and scratch never needs one.
func (p *planner) download(stage imagefile.Stage, digest string) bool {
	_, onStage := p.ends[stage.Base]

	return !onStage && stage.Base != scratch && digest == ""
}

func (p *planner) start(stage imagefile.Stage) (key string, digest string, err error) {
	if end, onStage := p.ends[stage.Base]; onStage {
		return end, "", nil
	}

	digest, err = baseDigest(stage, p.bases)
	if err != nil {
		return "", "", err
	}

	// an image nobody has fetched yet has no digest, and a key must not
	// stand for bytes nobody has seen
	if stage.Base != scratch && digest == "" {
		return "", "", nil
	}

	return baseKey(p.agent, stage.Base, digest), digest, nil
}

// steps also hands back where the stage's root file system ends, which for
// a stage without steps is where it started.
func (p *planner) steps(from string, stage imagefile.Stage) ([]Step, string, error) {
	var steps []Step
	rootfs := from
	var artifacts []string
	for _, instruction := range stage.Instructions {
		builtOn := []string{rootfs}
		if _, isCheck := instruction.(imagefile.Check); isCheck {
			// a copy, the steps must not share one list that still grows
			builtOn = slices.Clone(artifacts)
		}

		step, err := p.step(builtOn, instruction)
		if err != nil {
			return nil, "", err
		}

		steps = append(steps, step)
		switch instruction.(type) {
		case imagefile.Output, imagefile.Check:
			artifacts = append(artifacts, step.Key)
		default:
			rootfs = step.Key
		}
	}

	return steps, rootfs, nil
}

// step chains a step to what it is built on, so a change early in a stage
// reaches every key after it. The three lists are hashed apart, so that
// none can run into another.
func (p *planner) step(builtOn []string, instruction imagefile.Instruction) (Step, error) {
	files, err := p.files(instruction)
	if err != nil {
		return Step{}, err
	}

	reads := p.reads(instruction, files)

	// a step built on something without a key has none either
	key := ""
	if known(builtOn) && known(reads) {
		key = hashed([]string{hashed(builtOn), hashed(words(instruction)), hashed(reads)})
	}

	return Step{Instruction: instruction, Key: key, BuiltOn: builtOn, Files: files}, nil
}

func known(keys []string) bool {
	return !slices.Contains(keys, "")
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
		return p.stageReads(step)
	}

	var digests []string
	for _, file := range files {
		digests = append(digests, file.Digest)
	}

	return digests
}

// stageReads are one key for each source: that of the output it names, or
// else where the stage ends.
func (p *planner) stageReads(step imagefile.Copy) []string {
	var read []string
	for _, source := range step.Sources {
		key, isOutput := p.outputs[step.From][source]
		if !isOutput {
			key = p.ends[step.From]
		}

		read = append(read, key)
	}

	return read
}

func outputKeys(steps []Step) map[string]string {
	keys := map[string]string{}
	for _, step := range steps {
		if output, isOutput := step.Instruction.(imagefile.Output); isOutput {
			keys[output.Name] = step.Key
		}
	}

	return keys
}
