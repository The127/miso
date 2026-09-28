package protocol

import "io"

// Entries are the entries of a copy one after the other, each with its
// content, and io.EOF after the last.
type Entries interface {
	Next() (Entry, io.Reader, error)
}
