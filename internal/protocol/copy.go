package protocol

// Copy asks the agent to put sources of the build context into the image
// on top of layers and keep what that changes as the layer of the key. The
// files follow only once the agent asks for them.
type Copy struct {
	Key string

	// the keys of the layers below, lowest first
	Layers []string

	// paths in the build context, as the build file names them
	Sources []string

	// what the plan says of each source, in the same order. What arrives
	// is kept only when it is what was planned
	Digests []string

	// the absolute path in the image
	Destination string
}

func (c Copy) into(e *envelope) { e.Copy = &c }
