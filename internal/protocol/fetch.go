package protocol

// Fetch asks the agent for the disk image kept as the output of the key, or
// for one of the other files kept with it.
type Fetch struct {
	Key string

	// the name of the file, empty for the disk image
	File string
}

func (f Fetch) into(e *envelope) { e.Fetch = &f }
