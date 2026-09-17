package plan

import (
	"crypto/sha256"
	"encoding/hex"

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
	var kind, text string
	switch step := instruction.(type) {
	case imagefile.Run:
		kind, text = "RUN", step.Command
	case imagefile.Env:
		kind, text = "ENV", step.Key+"="+step.Value
	case imagefile.Check:
		kind, text = "CHECK", step.Command
	}

	sum := sha256.Sum256([]byte(parent + "\n" + kind + "\n" + text))
	return hex.EncodeToString(sum[:])
}
