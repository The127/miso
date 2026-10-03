package build

import (
	"errors"
	"fmt"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// ErrNoStepAtLine is a line of the build file that holds no step.
var ErrNoStepAtLine = errors.New("no step at this line")

// ErrNoStage is a build file with nothing to stand a shell on.
var ErrNoStage = errors.New("no stage")

// Before are the requests that bring the layers to where the step on the
// line of the build file would run, and a shell on them, in place of that
// step.
func Before(planned plan.Plan, network protocol.Network, line int) ([]Request, error) {
	requests, stopped, err := requestsUntil(planned, network, shellAt{line: line})
	if err != nil {
		return nil, err
	}

	if !stopped {
		return nil, fmt.Errorf("line %d: %w", line, ErrNoStepAtLine)
	}

	return requests, nil
}

// After are the requests that bring the layers to where the last stage
// ends, and a shell on them.
func After(planned plan.Plan, network protocol.Network) ([]Request, error) {
	requests, stopped, err := requestsUntil(planned, network, shellAt{end: true})
	if err != nil {
		return nil, err
	}

	if !stopped {
		return nil, ErrNoStage
	}

	return requests, nil
}

// shellOf is a shell on the layers and the environment a step would run on,
// with the network the step would have.
func shellOf(step plan.Step, under rootfs, network protocol.Network) Request {
	line, written := imagefile.Written(step.Instruction)

	return shellRequest(under, networkFor(step.Instruction, network), line, "shell before "+written)
}

// shellRequest is a shell on a root file system. A shell at no line of the
// build file has none to be named at.
func shellRequest(under rootfs, network *protocol.Network, line int, written string) Request {
	return Request{Line: line, Written: written, Message: protocol.Shell{Layers: under.layers, Env: under.env, Network: network}}
}

// IsFrom tells whether the line of the build file is the FROM of a stage,
// which brings an image in and is no step a shell can stand before.
func IsFrom(requests []Request, line int) bool {
	for _, request := range requests {
		if _, isImport := request.Message.(protocol.Import); isImport && request.Line == line {
			return true
		}
	}

	return false
}
