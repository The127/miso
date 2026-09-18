package protocol

import "errors"

// ErrAnotherAgent is a message sent by a miso of another version. Its keys
// may name layers this side would build differently.
var ErrAnotherAgent = errors.New("message from another agent")
