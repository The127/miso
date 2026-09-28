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

// StartOn starts the machine as if the host's KVM device were kvm.
func StartOn(ctx context.Context, d Driver, machine Machine, kvm string) (*VM, error) {
	return d.start(ctx, machine, kvm)
}
