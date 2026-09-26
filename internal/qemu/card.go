package qemu

// Card is the machine's network card on QEMU's user network. The agent finds
// it by its MAC.
type Card struct {
	MAC string
}

// network puts the card on QEMU's user network, so the machine needs
// nothing of the host's network set up.
func network(machine Machine) []string {
	if machine.Card == nil {
		return nil
	}

	return []string{
		"-netdev", "user,id=card",
		"-device", "virtio-net-pci,netdev=card,mac=" + machine.Card.MAC,
	}
}
