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
