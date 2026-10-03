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

// Before are the requests that bring the layers to where the step on the
// line of the build file would run, and a shell on them, in place of that
// step.
func Before(planned plan.Plan, network protocol.Network, line int) ([]Request, error) {
	requests, stopped, err := requestsUntil(planned, network, line)
	if err != nil {
		return nil, err
	}

	if !stopped {
		return nil, fmt.Errorf("line %d: %w", line, ErrNoStepAtLine)
	}

	return requests, nil
}

// shellOf is a shell on the layers and the environment a step would run on,
// with the network the step would have.
func shellOf(step plan.Step, under rootfs, network protocol.Network) Request {
	shell := protocol.Shell{Layers: under.layers, Env: under.env, Network: networkFor(step.Instruction, network)}
	line, written := imagefile.Written(step.Instruction)

	return Request{Line: line, Written: "shell before " + written, Message: shell}
}
