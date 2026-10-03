package protocol

import "io"

// Resize tells that the terminal of the user changed its size.
type Resize struct {
	Rows, Cols uint16
}

func (r Resize) into(e *envelope) { e.Resize = &r }

// Terminal is the terminal of the user as the agent's shell sees it.
type Terminal struct {
	// what the user types
	In io.Reader

	// the sizes the terminal changes to, nil when it never does
	Resized <-chan Resize
}

// Latest puts a size in the channel, which has room for one, in place of the
// one waiting there. Only the latest size matters, and the caller is the
// only sender, or it would block.
func Latest(resized chan Resize, size Resize) {
	select {
	case <-resized:
	default:
	}

	resized <- size
}

// Follow applies each size the terminal changes to, until over is closed or
// apply says to stop.
func (t Terminal) Follow(over <-chan struct{}, apply func(Resize) (more bool)) {
	for {
		select {
		case size := <-t.Resized:
			if !apply(size) {
				return
			}
		case <-over:
			return
		}
	}
}
