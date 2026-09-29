package protocol

import "io"

// Pieces send a fetched disk to the host: its length first, then each
// stretch of it that holds data.
type Pieces interface {
	Length(size int64) error
	Piece(piece Piece, content io.Reader) error
}

// sending sends the pieces of a disk over a connection.
type sending struct {
	conn *Conn
}

func (s sending) Length(size int64) error {
	return s.conn.Send(Length{Size: size})
}

func (s sending) Piece(piece Piece, content io.Reader) error {
	return s.conn.SendPiece(piece, content)
}
