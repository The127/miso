package protocol

// Length is how many bytes a fetched disk takes, holes counted. It comes
// before the pieces, so a disk that ends in a hole keeps its end.
type Length struct {
	Size int64
}

func (l Length) into(e *envelope) { e.Length = &l }
