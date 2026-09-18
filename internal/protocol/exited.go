package protocol

// Exited tells that a command ended with the code it exited with.
type Exited struct {
	Code int
}

func (Exited) message() {}
