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
	var keys [][]string
	last := map[string]string{}
	for _, stage := range stages {
		var stageKeys []string
		parent := hashed([]string{agent, stage.Base, bases.Digest(stage.Base)})
		for _, instruction := range stage.Instructions {
			var err error
			parent, err = stepKey(parent, instruction, last, context)
			if err != nil {
				return nil, err
			}

			stageKeys = append(stageKeys, parent)
		}

		keys = append(keys, stageKeys)
		if stage.Name != "" {
			last[stage.Name] = parent
		}
	}

	return keys, nil
}

// stepKey chains a step to its parent, so a change early in a stage reaches
// every key after it.
func stepKey(parent string, instruction imagefile.Instruction, last map[string]string, context Context) (string, error) {
	fields := []string{parent}
	switch step := instruction.(type) {
	case imagefile.Run:
		fields = append(fields, "RUN", step.Command)
	case imagefile.Env:
		fields = append(fields, "ENV", step.Key, step.Value)
	case imagefile.Copy:
		copied, err := copyFields(step, last, context)
		if err != nil {
			return "", err
		}

		fields = append(fields, copied...)
	case imagefile.Output:
		fields = append(fields, "OUTPUT", step.Kind)
		for _, name := range slices.Sorted(maps.Keys(step.Options)) {
			fields = append(fields, name, step.Options[name])
		}
	case imagefile.Check:
		fields = append(fields, "CHECK", step.Command)
	}

	return hashed(fields), nil
}

func copyFields(step imagefile.Copy, last map[string]string, context Context) ([]string, error) {
	fields := []string{"COPY", last[step.From]}
	for _, source := range step.Sources {
		fields = append(fields, source)
		if step.From != "" {
			continue
		}

		digest, err := context.Digest(source)
		if err != nil {
			return nil, at(step.Line, fmt.Errorf("COPY %s: %w", source, err))
		}

		fields = append(fields, digest)
	}

	return append(fields, step.Destination), nil
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
