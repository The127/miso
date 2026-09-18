package protocol

import "errors"

// ErrAgentFailed is the agent failing at its own work. The build file is
// not at fault.
var ErrAgentFailed = errors.New("agent failed")

// Failed tells that the agent could not do what it was asked. It is never a
// command that exited with an error, that is Exited.
type Failed struct {
	Reason string
}

func (Failed) message() {}
