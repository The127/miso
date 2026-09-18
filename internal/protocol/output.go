package protocol

// Output is a piece of what a command wrote, as it wrote it. It is bytes
// and not lines, because progress bars redraw with a carriage return.
type Output struct {
	Bytes []byte
}

func (Output) message() {}
