package qemu

import "errors"

// board is the kind of machine QEMU runs. What its virtio devices are called
// depends on the bus the board has for them.
type board struct {
	name string

	// The QEMU binary that runs this board.
	binary string

	// The end of the name of a virtio device of this board.
	bus string

	// The board has a firmware even when it boots a kernel, which brings the
	// SMBIOS table that systemd reads its credentials from. One without hands
	// them over on the kernel command line.
	firmware bool

	// The board boots a UEFI firmware from its disks.
	bootsFirmware bool
}

var (
	// q35 has virtio devices on a PCI bus.
	q35 = board{name: "q35", binary: "qemu-system-x86_64", bus: "-pci", firmware: true, bootsFirmware: true}

	// microvm has no PCI bus, so its devices and its power off go through
	// ACPI. Whatever the host's QEMU defaults to, ACPI is on. Linux boots it
	// with null_legacy_pic, so the i8259 stays at QEMU's reset state, vector 0
	// and unmasked. A timer interrupt pending at the first sti then dies as a
	// divide error. The PIC is always off, because the race hits only some boots.
	microvm = board{name: "microvm,acpi=on,pic=off", binary: "qemu-system-x86_64", bus: "-device"}

	// virt is the board of arm64. It has virtio devices on a PCI bus and boots
	// a firmware or a kernel.
	virt = board{name: "virt", binary: "qemu-system-aarch64", bus: "-pci", bootsFirmware: true}
)

const arm64 = "arm64"

// boardOf is the board the machine runs on.
func boardOf(machine Machine) board {
	if machine.Arch == arm64 {
		return virt
	}

	if machine.Microvm {
		return microvm
	}

	return q35
}

// device is the name of a virtio device of a kind on this board.
func (b board) device(kind string) string {
	return kind + b.bus
}

// boardless refuses a machine that no board fits.
func boardless(machine Machine) error {
	if machine.Arch == arm64 && machine.Microvm {
		return errors.New("an arm64 machine cannot be a microvm, QEMU has that board only for x86")
	}

	return nil
}
