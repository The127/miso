package protocol

import "io"

// DiskFile is where a fetched disk goes.
type DiskFile interface {
	io.WriterAt
	Truncate(size int64) error
}

// AskFetch has the agent send the disk of a key and writes it into disk at
// its length, with the pieces that hold data each at its offset.
func (c *Conn) AskFetch(request Fetch, disk DiskFile, out io.Writer) error {
	return c.ask(request, nil, disk, out)
}
