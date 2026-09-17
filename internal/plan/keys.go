package plan

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/The127/miso/internal/imagefile"
)

func runKey(run imagefile.Run) string {
	sum := sha256.Sum256([]byte(run.Command))
	return hex.EncodeToString(sum[:])
}
