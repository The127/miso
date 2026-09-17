package plan

import (
	"errors"
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// ErrDuplicateStage is a stage name that an earlier FROM already took.
var ErrDuplicateStage = errors.New("stage name taken")

func checkName(stage imagefile.Stage, known map[string]bool) error {
	if stage.Name != "" && known[stage.Name] {
		return at(stage.Line, fmt.Errorf("stage %s: %w", stage.Name, ErrDuplicateStage))
	}

	return nil
}
