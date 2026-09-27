package builder

// VM is the builder VM the agent runs in.
type VM interface {
	// Done is closed once the VM has stopped.
	Done() <-chan struct{}

	// Err is why the VM stopped, once Done is closed.
	Err() error
}
