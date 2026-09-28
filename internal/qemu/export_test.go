package qemu

import "context"

var (
	Arguments = arguments
	Accel     = accel
	Vsock     = vsock
	Claim     = claim
	OpenVsock = openVsock
	TakeCID   = takeCID
)

// StartOn starts the machine as if the host's KVM device were kvm and its
// vsock device were vsock.
func StartOn(ctx context.Context, d Driver, machine Machine, kvm, vsock string) (*VM, error) {
	return d.start(ctx, machine, kvm, vsock)
}
