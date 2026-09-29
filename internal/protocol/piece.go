package protocol

import "errors"

// ErrNegativeOffset is a piece that says it lies before the start of its
// disk.
var ErrNegativeOffset = errors.New("negative offset")

// Piece is a stretch of a fetched disk that holds data, at its offset from
// the start of the disk. Its raw bytes follow it. What no piece covers is a
// hole.
type Piece struct {
	Offset int64
	Size   int64
}

func (p Piece) into(e *envelope) { e.Piece = &p }
