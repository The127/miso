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

// baseDigests are empty for scratch, it has nothing to look up.
func baseDigests(stage imagefile.Stage, bases Bases) ([]string, error) {
	if stage.Base == scratch {
		return nil, nil
	}

	digest, err := bases.Digest(stage.Base)
	if err != nil {
		return nil, at(stage.Line, fmt.Errorf("FROM %s: %w", stage.Base, err))
	}

	return []string{digest}, nil
}

// baseKey holds the agent too, because a changed agent may build the same
// step differently.
func baseKey(agent string, base string, digests []string) string {
	return hashed(append([]string{agent, base}, digests...))
}
