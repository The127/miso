package plan

import (
	"errors"
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// ErrUnknownStage is a reference to a stage that no earlier FROM named.
var ErrUnknownStage = errors.New("unknown stage")

// Validate checks that the stages make sense together.
func Validate(stages []imagefile.Stage) error {
	known := map[string]bool{}
	for _, stage := range stages {
		for _, instruction := range stage.Instructions {
			copied, isCopy := instruction.(imagefile.Copy)
			if isCopy && copied.From != "" && !known[copied.From] {
				return &imagefile.Error{
					Line: copied.Line,
					Err:  fmt.Errorf("COPY --from=%s: %w", copied.From, ErrUnknownStage),
				}
			}
		}

		known[stage.Name] = true
	}

	return nil
}
