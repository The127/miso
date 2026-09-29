package plan

import (
	"errors"
	"fmt"
	"strings"

	"github.com/The127/miso/internal/imagefile"
)

// ErrUnknownStage is a reference to a stage that no earlier FROM named.
var ErrUnknownStage = errors.New("unknown stage")

// ErrUnknownOutput is a copy from a stage of a name that is no OUTPUT of
// that stage. A path in the stage's root file system starts with a slash.
var ErrUnknownOutput = errors.New("unknown output")

func checkReferences(stage imagefile.Stage, known map[string]map[string]bool) error {
	for _, instruction := range stage.Instructions {
		if step, isCopy := instruction.(imagefile.Copy); isCopy {
			if err := checkCopyFrom(step, known); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkCopyFrom checks that a copy from a stage names an earlier one, and of
// it only outputs it has.
func checkCopyFrom(copied imagefile.Copy, known map[string]map[string]bool) error {
	if copied.From == "" {
		return nil
	}

	outputs, isStage := known[copied.From]
	if !isStage {
		return at(copied.Line, fmt.Errorf("COPY --from=%s: %w", copied.From, ErrUnknownStage))
	}

	for _, source := range copied.Sources {
		if !strings.HasPrefix(source, "/") && !outputs[source] {
			return at(copied.Line, fmt.Errorf("COPY --from=%s %s: %w", copied.From, source, ErrUnknownOutput))
		}
	}

	return nil
}
