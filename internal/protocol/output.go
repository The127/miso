package protocol

// Output is a piece of what a command wrote, as it wrote it. It is bytes
// and not lines, because progress bars redraw with a carriage return.
type Output struct {
	Stream Stream
	Bytes  []byte
}

func (Output) message() {}

// Stream is where a command wrote its output.
type Stream string

// Stderr is a command's standard error.
const Stderr Stream = "stderr"
