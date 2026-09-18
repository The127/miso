package protocol

import "errors"

// ErrUnknownMessage is a line that holds no message this side knows.
var ErrUnknownMessage = errors.New("unknown message")

// Message is anything miso and its agent send each other.
type Message interface {
	into(e *envelope)
}

// envelope is a message on the wire, sent by the agent it names. Exactly
// one other field is set, and that field names what the line holds.
type envelope struct {
	Agent  string
	Run    *Run    `json:",omitempty"`
	Output *Output `json:",omitempty"`
	Done   *Done   `json:",omitempty"`
	Exited *Exited `json:",omitempty"`
	Failed *Failed `json:",omitempty"`
}

// open is the message the envelope holds, or nil when it holds none this
// side knows.
func (e envelope) open() Message {
	switch {
	case e.Run != nil:
		return *e.Run
	case e.Output != nil:
		return *e.Output
	case e.Done != nil:
		return *e.Done
	case e.Exited != nil:
		return *e.Exited
	case e.Failed != nil:
		return *e.Failed
	}

	return nil
}
