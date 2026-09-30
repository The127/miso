package protocol

import "time"

// Prune asks the agent to remove the layers that no build used for longer
// than OlderThan, and then the least recently used until what stays takes
// no more than KeepStorage bytes. A limit that is zero is not set.
type Prune struct {
	OlderThan   time.Duration
	KeepStorage int64
}

func (p Prune) into(e *envelope) { e.Prune = &p }
