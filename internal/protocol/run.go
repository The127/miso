package protocol

// Run asks the agent to run a command on top of layers and keep what it
// changes as the layer of the key.
type Run struct {
	Key string

	// the keys of the layers below, lowest first
	Layers  []string
	Command string
}

func (Run) message() {}
