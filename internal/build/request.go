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

// ErrUnknownKind is an output of a kind miso cannot make.
var ErrUnknownKind = errors.New("unknown kind of output")

// ErrUnknownOption is an option an output of its kind does not take.
var ErrUnknownOption = errors.New("unknown option")

// ErrOptionTakesNoValue is an option that only says yes, given a value.
var ErrOptionTakesNoValue = errors.New("option takes no value")

// ErrUnknownFormat is a file system a rootfs cannot be.
var ErrUnknownFormat = errors.New("unknown file system")

// Request is what the agent is asked for a step of the build file.
type Request struct {
	Line    int
	Written string
	Message protocol.Message

	// the name of the file a fetch writes, taken from the plan and never
	// from the agent
	Output string

	// what a fetched disk must pass when booted
	Checks []imagefile.Check

	// the sum of the fetched file is listed in the SHA256SUMS of its
	// directory
	Listed bool

	// the fetched disk is an ISO, booted from a CD when checked
	CD bool

	// what a fetched rootfs is booted with when checked
	Boot *BootFiles

	// the checks of a later output boot with the fetched file, so it is
	// fetched when no output directory is given too
	Needed bool
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
			requests = append(requests, importOf(stage))
		}

		// the plan puts an OUTPUT before every CHECK of its stage
		fetched := -1
		var inputs diskInputs

		boots := newBootingOutputs()
		fetchedKind := ""
		for _, step := range stage.Steps {
			under := roots[step.BuiltOn[0]]

			inputs.add(step.Instruction)

			message, err := messageOf(step, under, ends, network, inputs)
			if err != nil {
				return nil, err
			}

			if message != nil {
				line, written := imagefile.Written(step.Instruction)
				requests = append(requests, Request{Line: line, Written: written, Message: message})
			}

			if output, isOutput := step.Instruction.(imagefile.Output); isOutput {
				requests = append(requests, fetchesOf(step, output, inputs.partitions)...)
				fetched = len(requests) - 1
				fetchedKind = output.Kind
				boots.add(output, fetched, inputs.cmdline)
			}

			if check, isCheck := step.Instruction.(imagefile.Check); isCheck {
				requests[fetched].Checks = append(requests[fetched].Checks, check)

				if fetchedKind == plan.KindRootfs {
					requests[fetched].Boot = boots.bootFiles(requests)
				}
			}

			roots[step.Key] = under.after(step)
		}

		ends[stage.Name] = roots[stage.End]
	}

	return requests, nil
}

// importOf brings the image a stage is on into the cache.
func importOf(stage plan.Stage) Request {
	return Request{Line: stage.Line, Written: "FROM " + stage.Base, Message: protocol.Import{Key: stage.BaseKey, Digest: stage.BaseDigest}}
}

// messageOf is what the agent is asked for a step on the root file system
// under it, or nothing for a step the agent has no part in.
func messageOf(step plan.Step, under rootfs, ends map[string]rootfs, network protocol.Network, inputs diskInputs) (protocol.Message, error) {
	switch instruction := step.Instruction.(type) {
	case imagefile.Run:
		run := protocol.Run{Key: step.Key, Layers: under.layers, Env: under.env, Command: instruction.Command}
		if !instruction.Offline {
			run.Network = &network
		}

		return run, nil
	case imagefile.Copy:
		return copyRequest(step, instruction, under, ends[instruction.From])
	case imagefile.Output:
		return outputRequest(step, instruction, under, ends[plan.BuiltinTools], inputs)
	}

	return nil, nil
}
