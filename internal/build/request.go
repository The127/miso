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

// Requests are what the agent is asked, in the order it is asked.
func Requests(planned plan.Plan) ([]protocol.Message, error) {
	if len(planned.Downloads) > 0 {
		return nil, fmt.Errorf("%s: %w", strings.Join(planned.Downloads, ", "), ErrNotFetched)
	}

	var requests []protocol.Message
	roots := map[string]rootfs{}
	for _, stage := range planned.Stages {
		// only a stage on an image has a digest. A stage on an earlier stage
		// carries on where that one ended, and scratch needs no entry, a
		// missing key is already nothing
		_, seen := roots[stage.BaseKey]
		if !seen && stage.BaseDigest != "" {
			roots[stage.BaseKey] = rootfs{layers: []string{stage.BaseKey}}
			requests = append(requests, protocol.Import{Key: stage.BaseKey, Digest: stage.BaseDigest})
		}

		for _, step := range stage.Steps {
			under := roots[step.BuiltOn[0]]
			if run, isRun := step.Instruction.(imagefile.Run); isRun {
				requests = append(requests, protocol.Run{Key: step.Key, Layers: under.layers, Env: under.env, Command: run.Command, Offline: run.Offline})
			}

			roots[step.Key] = under.after(step)
		}
	}

	return requests, nil
}
