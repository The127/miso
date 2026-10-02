package qemu

// board is the kind of machine QEMU runs. What its virtio devices are called
// depends on the bus the board has for them.
type board struct {
	name string

	// The end of the name of a virtio device of this board.
	bus string
}

var (
	// q35 has virtio devices on a PCI bus.
	q35 = board{name: "q35", bus: "-pci"}

	// microvm has no PCI bus, so its devices and its power off go through
	// ACPI. Whatever the host's QEMU defaults to, ACPI is on.
	microvm = board{name: "microvm,acpi=on", bus: "-device"}
)

// boardOf is the board the machine runs on.
func boardOf(machine Machine) board {
	if machine.Microvm {
		return microvm
	}

	return q35
}

// device is the name of a virtio device of a kind on this board.
func (b board) device(kind string) string {
	return kind + b.bus
}
