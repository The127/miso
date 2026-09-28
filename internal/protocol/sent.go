package protocol

// Sent tells the agent that every entry of the copy has been sent.
type Sent struct{}

func (s Sent) into(e *envelope) { e.Sent = &s }
