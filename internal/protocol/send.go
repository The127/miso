package protocol

// Send asks miso for the entries of the copy being run, because its layer
// is not cached.
type Send struct{}

func (s Send) into(e *envelope) { e.Send = &s }
