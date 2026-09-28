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

// ErrOutputNotBuilt is a copy from an output of a stage. Outputs are not
// built yet, so there is nothing to copy.
var ErrOutputNotBuilt = errors.New("output not built yet")

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
	ends := map[string]rootfs{}
	for _, stage := range planned.Stages {
		// only a stage on an image has a digest. A stage on an earlier stage
		// carries on where that one ended, and scratch needs no entry, a
		// missing key is already nothing
		_, seen := roots[stage.BaseKey]
		if !seen && stage.BaseDigest != "" {
			roots[stage.BaseKey] = rootfs{layers: []string{stage.BaseKey}}
			requests = append(requests, Request{Line: stage.Line, Written: "FROM " + stage.Base, Message: protocol.Import{Key: stage.BaseKey, Digest: stage.BaseDigest}})
		}

		end := roots[stage.BaseKey]
		for _, step := range stage.Steps {
			under := roots[step.BuiltOn[0]]

			var message protocol.Message
			switch instruction := step.Instruction.(type) {
			case imagefile.Run:
				run := protocol.Run{Key: step.Key, Layers: under.layers, Env: under.env, Command: instruction.Command}
				if !instruction.Offline {
					run.Network = &network
				}

				message = run
			case imagefile.Copy:
				copying, err := copyRequest(step, instruction, under, ends[instruction.From])
				if err != nil {
					return nil, err
				}

				message = copying
			}

			if message != nil {
				line, written := imagefile.Written(step.Instruction)
				requests = append(requests, Request{Line: line, Written: written, Message: message})
			}

			roots[step.Key] = under.after(step)

			// outputs and checks leave the root file system unchanged, as the
			// plan has it
			switch step.Instruction.(type) {
			case imagefile.Output, imagefile.Check:
			default:
				end = roots[step.Key]
			}
		}

		ends[stage.Name] = end
	}

	return requests, nil
}
