package protocol

// Shell asks the agent for an interactive shell on top of layers. Nothing it
// changes is kept.
type Shell struct {
	// the keys of the layers below, lowest first
	Layers []string

	// as in Run
	Env []string

	// nil for a shell without network
	Network *Network
}

func (s Shell) into(e *envelope) { e.Shell = &s }
