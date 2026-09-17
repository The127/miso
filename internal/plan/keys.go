package plan

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/The127/miso/internal/imagefile"
)

// runKey chains a step to its parent, so a change early in a stage reaches
// every key after it.
func runKey(parent string, run imagefile.Run) string {
	sum := sha256.Sum256([]byte(parent + "\n" + run.Command))
	return hex.EncodeToString(sum[:])
}
