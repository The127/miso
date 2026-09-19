package protocol

// Run asks the agent to run a command on top of layers and keep what it
// changes as the layer of the key.
type Run struct {
	Key string

	// the keys of the layers below, lowest first
	Layers []string

	// KEY=VALUE as the build file set them, each key once where it was
	// first set. Defaults such as PATH are the agent's, so nothing of the
	// host's environment reaches a build
	Env     []string
	Command string

	// nil for a run without network
	Network *Network
}

func (r Run) into(e *envelope) { e.Run = &r }
