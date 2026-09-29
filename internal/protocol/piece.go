package protocol

import (
	"errors"
	"io"
)

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

// SendPiece writes a piece to the other side and its bytes right after it.
func (c *Conn) SendPiece(piece Piece, content io.Reader) error {
	if err := c.Send(piece); err != nil {
		return err
	}

	_, err := io.CopyN(c.w, content, piece.Size)

	return err
}
