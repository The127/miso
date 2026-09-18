package protocol

import "errors"

// ErrCommandFailed is a command of the build file that exited with an
// error. The build file is at fault, not miso.
var ErrCommandFailed = errors.New("command failed")

// Exited tells that a command ended with the code it exited with.
type Exited struct {
	Code int
}

func (x Exited) into(e *envelope) { e.Exited = &x }
