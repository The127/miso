package qemu

import (
	"os"

	"golang.org/x/sys/unix"
)

const kvmDevice = "/dev/kvm"

// kvmGetAPIVersion is KVM_GET_API_VERSION, _IO(0xAE, 0x00). Only KVM
// answers it, so a device that merely opens is not taken for KVM.
const kvmGetAPIVersion = 0xAE00

// accel runs the machine on KVM when the host's KVM answers, and else on
// TCG, which is much slower, and says why. On KVM the host's own CPU model
// hands the guest every feature the host has, the builder VM never moves
// to another host. TCG has no host CPU to hand on, its most capable model
// comes closest.
func accel(device string) ([]string, error) {
	tcg := []string{"-accel", "tcg", "-cpu", "max"}

	kvm, err := os.OpenFile(device, os.O_RDWR, 0)
	if err != nil {
		return tcg, err
	}

	defer func() { _ = kvm.Close() }()

	if _, err := unix.IoctlRetInt(int(kvm.Fd()), kvmGetAPIVersion); err != nil {
		return tcg, err
	}

	return []string{"-accel", "kvm", "-cpu", "host"}, nil
}
