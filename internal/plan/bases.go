package plan

import (
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// Bases knows the images a FROM can start a stage on.
type Bases interface {
	Digest(base string) (string, error)
}

// root is the key a stage starts from. The agent is part of it because a
// changed agent may build the same step differently.
func root(agent string, stage imagefile.Stage, bases Bases) (string, error) {
	fields := []string{agent, stage.Base}
	if stage.Base != "scratch" {
		digest, err := bases.Digest(stage.Base)
		if err != nil {
			return "", at(stage.Line, fmt.Errorf("FROM %s: %w", stage.Base, err))
		}

		fields = append(fields, digest)
	}

	return hashed(fields), nil
}
