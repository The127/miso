package plan

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/The127/miso/internal/imagefile"
)

// stepKey chains a step to its parent, so a change early in a stage reaches
// every key after it.
func stepKey(parent string, instruction imagefile.Instruction) string {
	var kind, text string
	switch step := instruction.(type) {
	case imagefile.Run:
		kind, text = "RUN", step.Command
	case imagefile.Check:
		kind, text = "CHECK", step.Command
	}

	sum := sha256.Sum256([]byte(parent + "\n" + kind + "\n" + text))
	return hex.EncodeToString(sum[:])
}
