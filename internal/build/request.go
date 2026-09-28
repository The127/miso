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

// Request is what the agent is asked for a step of the build file.
type Request struct {
	Line    int
	Written string
	Message protocol.Message
}

// Requests are what the agent is asked, in the order it is asked. A run
// with network gets the network the host set up for the builder.
func Requests(planned plan.Plan, network protocol.Network) ([]Request, error) {
	if len(planned.Downloads) > 0 {
		return nil, fmt.Errorf("%s: %w", strings.Join(planned.Downloads, ", "), ErrNotFetched)
	}

	var requests []Request
	roots := map[string]rootfs{}
	for _, stage := range planned.Stages {
		// only a stage on an image has a digest. A stage on an earlier stage
		// carries on where that one ended, and scratch needs no entry, a
		// missing key is already nothing
		_, seen := roots[stage.BaseKey]
		if !seen && stage.BaseDigest != "" {
			roots[stage.BaseKey] = rootfs{layers: []string{stage.BaseKey}}
			requests = append(requests, Request{Line: stage.Line, Written: "FROM " + stage.Base, Message: protocol.Import{Key: stage.BaseKey, Digest: stage.BaseDigest}})
		}

		for _, step := range stage.Steps {
			under := roots[step.BuiltOn[0]]
			if run, isRun := step.Instruction.(imagefile.Run); isRun {
				request := protocol.Run{Key: step.Key, Layers: under.layers, Env: under.env, Command: run.Command}
				if !run.Offline {
					request.Network = &network
				}

				line, written := imagefile.Written(run)
				requests = append(requests, Request{Line: line, Written: written, Message: request})
			}

			// a copy from a stage takes nothing from the build context
			if copying, isCopy := step.Instruction.(imagefile.Copy); isCopy && copying.From == "" {
				request := protocol.Copy{Key: step.Key, Layers: under.layers, Sources: copying.Sources, Destination: copying.Destination}
				for _, file := range step.Files {
					request.Digests = append(request.Digests, file.Digest)
				}

				requests = append(requests, Request{Message: request})
			}

			roots[step.Key] = under.after(step)
		}
	}

	return requests, nil
}
