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

// Vsock is the vsock device of whatever board the machine runs on.
func Vsock(cid uint32, fd int, machine Machine) []string {
	return vsock(cid, fd, boardOf(machine))
}

// Port is the virtio port of whatever board the machine runs on.
func Port(fd int, machine Machine) []string {
	return port(fd, boardOf(machine))
}
