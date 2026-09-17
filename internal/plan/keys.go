package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"slices"
	"strconv"

	"github.com/The127/miso/internal/imagefile"
)

// Keys are the cache keys of a build, one for each instruction of each
// stage.
func Keys(stages []imagefile.Stage) [][]string {
	var keys [][]string
	last := map[string]string{}
	for _, stage := range stages {
		var stageKeys []string
		parent := stage.Base
		for _, instruction := range stage.Instructions {
			parent = stepKey(parent, instruction, last)
			stageKeys = append(stageKeys, parent)
		}

		keys = append(keys, stageKeys)
		last[stage.Name] = parent
	}

	return keys
}

// stepKey chains a step to its parent, so a change early in a stage reaches
// every key after it.
func stepKey(parent string, instruction imagefile.Instruction, last map[string]string) string {
	fields := []string{parent}
	switch step := instruction.(type) {
	case imagefile.Run:
		fields = append(fields, "RUN", step.Command)
	case imagefile.Env:
		fields = append(fields, "ENV", step.Key, step.Value)
	case imagefile.Copy:
		fields = append(fields, "COPY", last[step.From])
		fields = append(fields, step.Sources...)
		fields = append(fields, step.Destination)
	case imagefile.Output:
		fields = append(fields, "OUTPUT", step.Kind)
		for _, name := range slices.Sorted(maps.Keys(step.Options)) {
			fields = append(fields, name, step.Options[name])
		}
	case imagefile.Check:
		fields = append(fields, "CHECK", step.Command)
	}

	// the length in front keeps "a b" apart from "a" and "b"
	hash := sha256.New()
	for _, field := range fields {
		hash.Write([]byte(strconv.Itoa(len(field)) + ":" + field))
	}

	return hex.EncodeToString(hash.Sum(nil))
}
