package protocol

// Fetch asks the agent for the disk image kept as the output of the key.
type Fetch struct {
	Key string
}

func (f Fetch) into(e *envelope) { e.Fetch = &f }
