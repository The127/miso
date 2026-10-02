package qemu

import "context"

var (
	Arguments = arguments
	Accel     = accel
	Claim     = claim
	TakeCID   = takeCID
)

// StartOn starts the machine as if the host's KVM device were kvm.
func StartOn(ctx context.Context, d Driver, machine Machine, kvm string) (*VM, error) {
	return d.start(ctx, machine, kvm)
}

// Vsock is the device of a machine on the microvm board or not.
func Vsock(cid uint32, fd int, onMicrovm bool) []string {
	return vsock(cid, fd, boardOf(Machine{Microvm: onMicrovm}))
}

// Port is the virtio port of a machine on the microvm board or not.
func Port(fd int, onMicrovm bool) []string {
	return port(fd, boardOf(Machine{Microvm: onMicrovm}))
}
