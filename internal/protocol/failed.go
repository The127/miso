package protocol

// Failed tells that the agent could not do what it was asked. It is never a
// command that exited with an error, that is Exited.
type Failed struct {
	Reason string
}

func (Failed) message() {}
