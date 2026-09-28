package copydigest

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

// format names the way a digest is made. A change to what goes into a
// digest, or how, gets a new name, so that no old digest can match a new one.
const format = "miso-context-1"

// Of is the digest of a source whose entries summed to sums, in walk order.
func Of(sums []string) string {
	return format + ":" + hashed(sums)
}

// hashed puts the length in front of every field, so that no field can
// run into the next.
func hashed(fields []string) string {
	hash := sha256.New()
	for _, field := range fields {
		hash.Write([]byte(strconv.Itoa(len(field)) + ":" + field))
	}

	return hex.EncodeToString(hash.Sum(nil))
}
