package build

import (
	"errors"
	"fmt"
	"strings"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// ErrNotFetched is a plan on a base that nobody has fetched yet. Its steps
// have no keys, so nothing can be asked for them.
var ErrNotFetched = errors.New("not fetched yet")

// Requests are what the agent is asked, one per step it carries out.
func Requests(planned plan.Plan) ([]protocol.Run, error) {
	if len(planned.Downloads) > 0 {
		return nil, fmt.Errorf("%s: %w", strings.Join(planned.Downloads, ", "), ErrNotFetched)
	}

	var requests []protocol.Run
	roots := map[string]rootfs{}
	for _, stage := range planned.Stages {
		for _, step := range stage.Steps {
			// a key nobody left behind is a base
			under, seen := roots[step.BuiltOn[0]]
			if !seen {
				under = rootfs{layers: []string{step.BuiltOn[0]}}
			}

			if run, isRun := step.Instruction.(imagefile.Run); isRun {
				requests = append(requests, protocol.Run{Key: step.Key, Layers: under.layers, Env: under.env, Command: run.Command})
			}

			roots[step.Key] = under.after(step)
		}
	}

	return requests, nil
}
