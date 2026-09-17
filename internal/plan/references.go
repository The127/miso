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
		copied, isCopy := instruction.(imagefile.Copy)
		if !isCopy || copied.From == "" {
			continue
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
	}

	return nil
}
