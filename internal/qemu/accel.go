package qemu

import "os"

const kvmDevice = "/dev/kvm"

// accel runs the machine on KVM when the host lets it open the device, and
// else on TCG, which is much slower, and says why. On KVM the host's own
// CPU model hands the guest every feature the host has, the builder VM
// never moves to another host. TCG has no host CPU to hand on, its most
// capable model comes closest.
func accel(device string) ([]string, error) {
	kvm, err := os.OpenFile(device, os.O_RDWR, 0)
	if err != nil {
		return []string{"-accel", "tcg", "-cpu", "max"}, err
	}

	_ = kvm.Close()

	return []string{"-accel", "kvm", "-cpu", "host"}, nil
}
