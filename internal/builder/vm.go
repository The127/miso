package builder

import "fmt"

// VM is the builder VM the agent runs in.
type VM interface {
	// Done is closed once the VM has stopped.
	Done() <-chan struct{}

	// Err is why the VM stopped, once Done is closed.
	Err() error

	// WithoutKVM is whether the VM runs on a host without KVM, much slower.
	WithoutKVM() bool
}

// stopped says that the VM stopped and when, and why if it says.
func stopped(vm VM, when string) error {
	if vm.Err() == nil {
		return fmt.Errorf("the builder VM stopped %s", when)
	}

	return fmt.Errorf("the builder VM stopped %s: %w", when, vm.Err())
}
