package plan

import (
	"errors"
	"fmt"
	"strings"

	"github.com/The127/miso/internal/imagefile"
)

// Tools is the option of an OUTPUT that names the stage bringing the tools
// to make it.
const Tools = "tools"

// ErrNotAFileName is an output name that is more than the name of a file,
// so the host would write it somewhere else than where outputs go.
var ErrNotAFileName = errors.New("not a plain file name")

// ErrDuplicateOutput is a file name that an earlier OUTPUT already writes.
var ErrDuplicateOutput = errors.New("file name taken")

func checkOutputs(stage imagefile.Stage, written map[string]bool) error {
	for _, instruction := range stage.Instructions {
		output, isOutput := instruction.(imagefile.Output)
		if !isOutput {
			continue
		}

		if strings.Contains(output.Name, "/") {
			return at(output.Line, fmt.Errorf("OUTPUT %q: %w", output.Name, ErrNotAFileName))
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
