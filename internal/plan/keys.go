package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/The127/miso/internal/imagefile"
)

// Keys are the cache keys of a build, one for each instruction of each
// stage.
func Keys(stages []imagefile.Stage, agent string, context Context, bases Bases) ([][]string, error) {
	planned, err := New(stages, agent, context, bases)
	if err != nil {
		return nil, err
	}

	var keys [][]string
	for _, stage := range planned.Stages {
		var stageKeys []string
		for _, step := range stage.Steps {
			stageKeys = append(stageKeys, step.Key)
		}

		keys = append(keys, stageKeys)
	}

	return keys, nil
}

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
	end := from
	for _, instruction := range stage.Instructions {
		step, err := p.step(end, instruction)
		if err != nil {
			return nil, "", err
		}

		steps = append(steps, step)
		end = step.Key
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

// words are what an instruction says.
func words(instruction imagefile.Instruction) []string {
	var found []string
	switch step := instruction.(type) {
	case imagefile.Run:
		found = []string{"RUN", step.Command}
	case imagefile.Env:
		found = []string{"ENV", step.Key, step.Value}
	case imagefile.Copy:
		// the stage's name stays out, its key is among the reads, so a
		// renamed stage keeps its cache
		found = []string{"COPY"}
		if step.From != "" {
			found = append(found, "--from")
		}

		found = append(found, step.Sources...)
		found = append(found, step.Destination)
	case imagefile.Output:
		found = []string{"OUTPUT", step.Kind}
		for _, name := range slices.Sorted(maps.Keys(step.Options)) {
			found = append(found, name, step.Options[name])
		}
	case imagefile.Check:
		found = []string{"CHECK", step.Command}
	}

	return found
}

func contextFiles(step imagefile.Copy, context Context) ([]File, error) {
	var files []File
	for _, source := range step.Sources {
		digest, err := context.Digest(source)
		if err != nil {
			return nil, at(step.Line, fmt.Errorf("COPY %s: %w", source, err))
		}

		files = append(files, File{Path: source, Digest: digest})
	}

	return files, nil
}

// hashed puts the length in front of every field, which keeps "a b" apart
// from "a" and "b".
func hashed(fields []string) string {
	hash := sha256.New()
	for _, field := range fields {
		hash.Write([]byte(strconv.Itoa(len(field)) + ":" + field))
	}

	return hex.EncodeToString(hash.Sum(nil))
}
