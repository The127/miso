package qemu

// accel runs the machine on KVM. The host's own CPU model hands the guest
// every feature the host has, the builder VM never moves to another host.
func accel() []string {
	return []string{"-enable-kvm", "-cpu", "host"}
}
