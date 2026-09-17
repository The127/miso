package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/The127/miso/internal/imagefile"
)

// Keys are the cache keys of a stage, one for each instruction.
func Keys(stage imagefile.Stage) []string {
	var keys []string
	parent := stage.Base
	for _, instruction := range stage.Instructions {
		parent = stepKey(parent, instruction)
		keys = append(keys, parent)
	}

	return keys
}

// stepKey chains a step to its parent, so a change early in a stage reaches
// every key after it.
func stepKey(parent string, instruction imagefile.Instruction) string {
	fields := []string{parent}
	switch step := instruction.(type) {
	case imagefile.Run:
		fields = append(fields, "RUN", step.Command)
	case imagefile.Env:
		fields = append(fields, "ENV", step.Key, step.Value)
	case imagefile.Copy:
		fields = append(fields, "COPY")
		fields = append(fields, step.Sources...)
		fields = append(fields, step.Destination)
	case imagefile.Output:
		fields = append(fields, "OUTPUT", step.Kind)
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
