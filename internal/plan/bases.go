package plan

import (
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// scratch is the base with nothing in it.
const scratch = "scratch"

// Bases knows the images a FROM can start a stage on.
type Bases interface {
	Digest(base string) (string, error)
}

// baseKey holds the agent too, because a changed agent may build the same
// step differently.
func baseKey(agent string, stage imagefile.Stage, bases Bases) (string, error) {
	fields := []string{agent, stage.Base}
	if stage.Base != scratch {
		digest, err := bases.Digest(stage.Base)
		if err != nil {
			return "", at(stage.Line, fmt.Errorf("FROM %s: %w", stage.Base, err))
		}

		fields = append(fields, digest)
	}

	return hashed(fields), nil
}
