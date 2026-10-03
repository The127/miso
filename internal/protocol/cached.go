package protocol

// Cached asks the agent which of the layers of the keys are finished, and
// is answered with those keys, one to a line.
type Cached struct {
	Keys []string
}

func (c Cached) into(e *envelope) { e.Cached = &c }
