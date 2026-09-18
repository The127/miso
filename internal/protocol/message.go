package protocol

import "errors"

// ErrUnknownMessage is a line that holds no message this side knows.
var ErrUnknownMessage = errors.New("unknown message")

// Message is anything miso and its agent send each other.
type Message interface {
	message()
}

// envelope is a message on the wire, sent by the agent it names. Exactly
// one other field is set, and that field names what the line holds.
type envelope struct {
	Agent  string
	Run    *Run    `json:",omitempty"`
	Output *Output `json:",omitempty"`
	Exited *Exited `json:",omitempty"`
}
