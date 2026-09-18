package protocol

// Import asks the agent to take the root file system of the base image with
// the digest and keep it as the layer of the key.
type Import struct {
	Key    string
	Digest string
}

func (i Import) into(e *envelope) { e.Import = &i }
