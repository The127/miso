package qemu

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

const kvmDevice = "/dev/kvm"

// ErrOtherArch is a machine of another architecture than the host's, which
// KVM cannot run.
var ErrOtherArch = errors.New("KVM runs only machines of the host's architecture")

// kvmGetAPIVersion is KVM_GET_API_VERSION, _IO(0xAE, 0x00). Only KVM
// answers it, so a device that merely opens is not taken for KVM.
const kvmGetAPIVersion = 0xAE00

// accel runs a machine of the host's architecture on KVM when the host's KVM
// answers, and else on TCG, which is much slower, and says why. On KVM the
// host's own CPU model hands the guest every feature the host has, the
// builder VM never moves to another host. TCG has no host CPU to hand on,
// its most capable model comes closest.
func accel(device, arch, host string) ([]string, error) {
	tcg := []string{"-accel", "tcg", "-cpu", "max"}
	if arch == "" {
		arch = "amd64"
	}

	if arch != host {
		return tcg, ErrOtherArch
	}

	kvm, err := os.OpenFile(device, os.O_RDWR, 0)
	if err != nil {
		return tcg, err
	}

	defer func() { _ = kvm.Close() }()

	if _, err := unix.IoctlRetInt(int(kvm.Fd()), kvmGetAPIVersion); err != nil {
		return tcg, &os.PathError{Op: "ask KVM", Path: device, Err: err}
	}

	return []string{"-accel", "kvm", "-cpu", "host"}, nil
}
