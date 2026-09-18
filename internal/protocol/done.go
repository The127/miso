package protocol

// Done tells that what was asked worked and its layer is saved.
type Done struct{}

func (Done) message() {}
