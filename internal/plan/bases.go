package plan

import (
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// Scratch is the base with nothing in it.
const Scratch = "scratch"

// Bases knows the images a FROM can start a stage on. An image that is
// known but not fetched yet has an empty digest and no error.
type Bases interface {
	Digest(base string) (string, error)
}

// baseDigest is empty for scratch, it has nothing to look up.
func baseDigest(stage imagefile.Stage, bases Bases) (string, error) {
	if stage.Base == Scratch {
		return "", nil
	}

	digest, err := bases.Digest(stage.Base)
	if err != nil {
		return "", at(stage.Line, fmt.Errorf("FROM %s: %w", stage.Base, err))
	}

	return digest, nil
}

// baseKey holds the agent too, because a changed agent may build the same
// step differently. The name stays out, so that two names for one image
// share its layers.
func baseKey(agent string, digest string) string {
	return hashed([]string{agent, digest})
}
