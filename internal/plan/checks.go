package plan

import (
	"errors"
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// ErrNothingToCheck is a CHECK with no OUTPUT before it in its stage.
var ErrNothingToCheck = errors.New("no OUTPUT before it")

func checkChecks(stage imagefile.Stage) error {
	produced := false
	for _, instruction := range stage.Instructions {
		switch step := instruction.(type) {
		case imagefile.Output:
			produced = true
		case imagefile.Check:
			if !produced {
				return at(step.Line, fmt.Errorf("CHECK: %w", ErrNothingToCheck))
			}
		}
	}

	return nil
}
