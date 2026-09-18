package protocol

// Message is anything miso and its agent send each other.
type Message interface {
	message()
}

// envelope is a message on the wire. Exactly one field is set, and that
// field names what the line holds.
type envelope struct {
	Run    *Run    `json:",omitempty"`
	Exited *Exited `json:",omitempty"`
}
