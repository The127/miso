package plan

import (
	"errors"
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// ErrDuplicateOutput is a file name that an earlier OUTPUT already writes.
var ErrDuplicateOutput = errors.New("file name taken")

func checkOutputs(stage imagefile.Stage, written map[string]bool) error {
	for _, instruction := range stage.Instructions {
		output, isOutput := instruction.(imagefile.Output)
		if !isOutput {
			continue
		}

		if written[output.Name] {
			return at(output.Line, fmt.Errorf("OUTPUT %s: %w", output.Name, ErrDuplicateOutput))
		}

		written[output.Name] = true
	}

	return nil
}

func outputNames(stage imagefile.Stage) map[string]bool {
	names := map[string]bool{}
	for _, instruction := range stage.Instructions {
		if output, isOutput := instruction.(imagefile.Output); isOutput {
			names[output.Name] = true
		}
	}

	return names
}
