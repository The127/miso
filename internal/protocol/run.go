package protocol

// Run asks the agent to run a command on top of layers and keep what it
// changes as the layer of the key.
type Run struct {
	Key string

	// the keys of the layers below, lowest first
	Layers []string

	// KEY=VALUE, in the order the build file set them
	Env     []string
	Command string
}

func (r Run) into(e *envelope) { e.Run = &r }
