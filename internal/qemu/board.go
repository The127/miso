package qemu

// board is the kind of machine QEMU runs. What its virtio devices are called
// depends on the bus the board has for them.
type board struct {
	name string

	// The end of the name of a virtio device of this board.
	bus string

	// The board has a firmware, which brings the SMBIOS table that
	// systemd reads its credentials from, and a drive for a CD. One without
	// boots a kernel, and its credentials go on the kernel command line.
	firmware bool
}

var (
	// q35 has virtio devices on a PCI bus.
	q35 = board{name: "q35", bus: "-pci", firmware: true}

	// microvm has no PCI bus, so its devices and its power off go through
	// ACPI. Whatever the host's QEMU defaults to, ACPI is on. Linux boots it
	// with null_legacy_pic, so the i8259 stays at QEMU's reset state, vector 0
	// and unmasked. A timer interrupt pending at the first sti then dies as a
	// divide error. The PIC is always off, because the race hits only some boots.
	microvm = board{name: "microvm,acpi=on,pic=off", bus: "-device"}

	// virt is the board of arm64. It has virtio devices on a PCI bus and boots
	// a firmware or a kernel.
	virt = board{name: "virt", bus: "-pci", firmware: true}
)

// boardOf is the board the machine runs on.
func boardOf(machine Machine) board {
	if machine.Arch == "arm64" {
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
