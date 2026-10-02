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
	Agent    string
	Run      *Run      `json:",omitempty"`
	Import   *Import   `json:",omitempty"`
	Copy     *Copy     `json:",omitempty"`
	Disk     *Disk     `json:",omitempty"`
	Rootfs   *Rootfs   `json:",omitempty"`
	BootPart *BootPart `json:",omitempty"`
	Fetch    *Fetch    `json:",omitempty"`
	Prune    *Prune    `json:",omitempty"`
	Entry    *Entry    `json:",omitempty"`
	Piece    *Piece    `json:",omitempty"`
	Length   *Length   `json:",omitempty"`
	Send     *Send     `json:",omitempty"`
	Sent     *Sent     `json:",omitempty"`
	Output   *Output   `json:",omitempty"`
	Done     *Done     `json:",omitempty"`
	Exited   *Exited   `json:",omitempty"`
	Failed   *Failed   `json:",omitempty"`
}

// open is the message the envelope holds, or nil when it holds none this
// side knows.
func (e envelope) open() Message {
	switch {
	case e.Run != nil:
		return *e.Run
	case e.Import != nil:
		return *e.Import
	case e.Copy != nil:
		return *e.Copy
	case e.Disk != nil:
		return *e.Disk
	case e.Rootfs != nil:
		return *e.Rootfs
	case e.BootPart != nil:
		return *e.BootPart
	case e.Fetch != nil:
		return *e.Fetch
	case e.Prune != nil:
		return *e.Prune
	case e.Entry != nil:
		return *e.Entry
	case e.Piece != nil:
		return *e.Piece
	case e.Length != nil:
		return *e.Length
	case e.Send != nil:
		return *e.Send
	case e.Sent != nil:
		return *e.Sent
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
