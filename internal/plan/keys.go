package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"slices"
	"strconv"

	"github.com/The127/miso/internal/imagefile"
)

// words are what an instruction says.
func words(instruction imagefile.Instruction) []string {
	var found []string
	switch step := instruction.(type) {
	case imagefile.Run:
		found = []string{"RUN"}
		if step.Offline {
			found = append(found, "--network=none")
		}

		found = append(found, step.Command)
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

// hashed puts the length in front of every field, which keeps "a b" apart
// from "a" and "b".
func hashed(fields []string) string {
	hash := sha256.New()
	for _, field := range fields {
		hash.Write([]byte(strconv.Itoa(len(field)) + ":" + field))
	}

	return hex.EncodeToString(hash.Sum(nil))
}
