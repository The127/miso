package qemu

import "fmt"

// Card is the machine's network card on QEMU's user network. The agent finds
// it by its MAC.
type Card struct {
	MAC  string
	IPv4 Family
}

// Family is how QEMU lays out one address family of its user network.
type Family struct {
	Prefix     string
	Gateway    string
	Nameserver string
}

// network puts the card on QEMU's user network, so the machine needs
// nothing of the host's network set up.
func network(machine Machine) []string {
	if machine.Card == nil {
		return nil
	}

	ipv4 := machine.Card.IPv4

	return []string{
		"-netdev", fmt.Sprintf("user,id=card,net=%s,host=%s,dns=%s", ipv4.Prefix, ipv4.Gateway, ipv4.Nameserver),
		"-device", "virtio-net-pci,netdev=card,mac=" + machine.Card.MAC,
	}
}
