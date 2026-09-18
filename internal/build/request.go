package build

import (
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// Requests are what the agent is asked, one per step it carries out.
func Requests(planned plan.Plan) []protocol.Run {
	var requests []protocol.Run
	for _, stage := range planned.Stages {
		for _, step := range stage.Steps {
			run := step.Instruction.(imagefile.Run)
			requests = append(requests, protocol.Run{Key: step.Key, Layers: step.BuiltOn, Command: run.Command})
		}
	}

	return requests
}
