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
	if err := Validate(stages); err != nil {
		return nil, err
	}

	var keys [][]string
	last := map[string]string{}
	for _, stage := range stages {
		from, err := start(stage, last, agent, bases)
		if err != nil {
			return nil, err
		}

		stageKeys, end, err := stepKeys(from, stage, last, context)
		if err != nil {
			return nil, err
		}

		keys = append(keys, stageKeys)
		if stage.Name != "" {
			last[stage.Name] = end
		}
	}

	return keys, nil
}

func start(stage imagefile.Stage, last map[string]string, agent string, bases Bases) (string, error) {
	if end, onStage := last[stage.Base]; onStage {
		return end, nil
	}

	digests, err := baseDigests(stage, bases)
	if err != nil {
		return "", err
	}

	return baseKey(agent, stage.Base, digests), nil
}

// stepKeys also hands back where the stage ends, which for a stage without
// steps is where it started.
func stepKeys(from string, stage imagefile.Stage, last map[string]string, context Context) ([]string, string, error) {
	var keys []string
	end := from
	for _, instruction := range stage.Instructions {
		key, err := stepKey(end, instruction, last, context)
		if err != nil {
			return nil, "", err
		}

		keys = append(keys, key)
		end = key
	}

	return keys, end, nil
}

// stepKey chains a step to its parent, so a change early in a stage reaches
// every key after it. Words and reads are hashed apart, so that neither
// list can run into the other.
func stepKey(parent string, instruction imagefile.Instruction, last map[string]string, context Context) (string, error) {
	read, err := reads(instruction, last, context)
	if err != nil {
		return "", err
	}

	return hashed([]string{parent, hashed(words(instruction)), hashed(read)}), nil
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

// reads are what a step takes from outside its own stage.
func reads(instruction imagefile.Instruction, last map[string]string, context Context) ([]string, error) {
	step, isCopy := instruction.(imagefile.Copy)
	if !isCopy {
		return nil, nil
	}

	if step.From != "" {
		return []string{last[step.From]}, nil
	}

	return contextDigests(step, context)
}

func contextDigests(step imagefile.Copy, context Context) ([]string, error) {
	var digests []string
	for _, source := range step.Sources {
		digest, err := context.Digest(source)
		if err != nil {
			return nil, at(step.Line, fmt.Errorf("COPY %s: %w", source, err))
		}

		digests = append(digests, digest)
	}

	return digests, nil
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
