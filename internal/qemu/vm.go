package qemu

// VM is a started builder VM.
type VM struct {
	cid  uint32
	done chan struct{}
}

// CID is where the host reaches the VM over vsock.
func (vm *VM) CID() uint32 {
	return vm.cid
}

// Done is closed once the VM has stopped.
func (vm *VM) Done() <-chan struct{} {
	return vm.done
}
