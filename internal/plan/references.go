package plan

import (
	"errors"
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// ErrUnknownStage is a reference to a stage that no earlier FROM named.
var ErrUnknownStage = errors.New("unknown stage")

func checkReferences(stage imagefile.Stage, known map[string]bool) error {
	for _, instruction := range stage.Instructions {
		copied, isCopy := instruction.(imagefile.Copy)
		if isCopy && copied.From != "" && !known[copied.From] {
			return at(copied.Line, fmt.Errorf("COPY --from=%s: %w", copied.From, ErrUnknownStage))
		}
	}

	return nil
}
