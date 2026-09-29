package protocol

import "io"

// AskFetch has the agent send the disk of a key and writes the pieces of it
// that hold data into disk, each at its offset.
func (c *Conn) AskFetch(request Fetch, disk io.WriterAt, out io.Writer) error {
	return c.ask(request, nil, disk, out)
}
