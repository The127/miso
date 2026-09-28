package protocol

// Copy asks the agent to put sources into the image on top of layers and
// keep what that changes as the layer of the key. Sources of the build
// context follow only once the agent asks for them. Sources of an earlier
// stage are already in its layers.
type Copy struct {
	Key string

	// the keys of the layers below, lowest first
	Layers []string

	// the stage the sources come from, and its layers lowest first. A stage
	// can have no layers, so only an empty name means the build context
	Stage string
	From  []string

	// paths in the build context or in that stage, as the build file names
	// them
	Sources []string

	// what the plan says of each source, in the same order. What arrives
	// is kept only when it is what was planned
	Digests []string

	// the absolute path in the image
	Destination string
}

func (c Copy) into(e *envelope) { e.Copy = &c }
