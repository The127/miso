package protocol

// Input is what the host's user typed in a shell, as bytes.
type Input struct {
	Bytes []byte
}

func (i Input) into(e *envelope) { e.Input = &i }
