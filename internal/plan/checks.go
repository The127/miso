package plan

import (
	"errors"
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// ErrNothingToCheck is a CHECK with no OUTPUT before it in its stage.
var ErrNothingToCheck = errors.New("no OUTPUT before it")

// ErrNotBooted is a CHECK after an output that is never booted.
var ErrNotBooted = errors.New("is never booted")

func checkChecks(stage imagefile.Stage) error {
	var last *imagefile.Output
	for _, instruction := range stage.Instructions {
		switch step := instruction.(type) {
		case imagefile.Output:
			last = &step
		case imagefile.Check:
			if last == nil {
				return at(step.Line, fmt.Errorf("CHECK: %w", ErrNothingToCheck))
			}

			if NeverBooted(last.Kind) {
				return at(step.Line, fmt.Errorf("CHECK: the %s %s %w", last.Kind, last.Name, ErrNotBooted))
			}
		}
	}

	return nil
}
