package build

import (
	"slices"
	"strings"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// Requests are what the agent is asked, one per step it carries out.
func Requests(planned plan.Plan) []protocol.Run {
	var requests []protocol.Run
	layers := map[string][]string{}
	envs := map[string][]string{}
	for _, stage := range planned.Stages {
		for _, step := range stage.Steps {
			below, seen := layers[step.BuiltOn[0]]
			if !seen {
				below = []string{step.BuiltOn[0]}
			}

			env := envs[step.BuiltOn[0]]
			if run, isRun := step.Instruction.(imagefile.Run); isRun {
				requests = append(requests, protocol.Run{Key: step.Key, Layers: below, Env: env, Command: run.Command})
			}

			layers[step.Key] = slices.Concat(below, []string{step.Key})
			envs[step.Key] = env
			if variable, isEnv := step.Instruction.(imagefile.Env); isEnv {
				layers[step.Key] = below
				set := slices.Clone(env)
				assignment := variable.Key + "=" + variable.Value
				at := slices.IndexFunc(set, func(earlier string) bool { return strings.HasPrefix(earlier, variable.Key+"=") })
				if at < 0 {
					set = append(set, assignment)
				} else {
					set[at] = assignment
				}

				envs[step.Key] = set
			}
		}
	}

	return requests
}
