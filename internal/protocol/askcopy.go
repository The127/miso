package protocol

import "io"

// Files are the entries of a copy on the host's side. They call send once
// for each entry, in order, with its content.
type Files func(send func(Entry, io.Reader) error) error

// AskCopy has the agent carry out a copy and sends it the files when it asks
// for them, which it does only when their layer is not cached.
func (c *Conn) AskCopy(request Copy, files Files, out io.Writer) error {
	return c.ask(request, files, out)
}
