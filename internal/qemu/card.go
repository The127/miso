package qemu

import "fmt"

// Card is the machine's network card on QEMU's user network. The agent finds
// it by its MAC.
type Card struct {
	MAC  string
	IPv4 Family
	IPv6 Family
}

// Family is how QEMU lays out one address family of its user network.
type Family struct {
	Prefix     string
	Gateway    string
	Nameserver string
}

// network puts the card on QEMU's user network, so the machine needs
// nothing of the host's network set up. A machine with no card gets none of
// QEMU's own either, which no fence of the agent's would guard.
func network(machine Machine) []string {
	if machine.Card == nil {
		return []string{"-nic", "none"}
	}

	ipv4, ipv6 := machine.Card.IPv4, machine.Card.IPv6

	return []string{
		"-netdev", fmt.Sprintf("user,id=card,net=%s,host=%s,dns=%s,ipv6-net=%s,ipv6-host=%s,ipv6-dns=%s",
			ipv4.Prefix, ipv4.Gateway, ipv4.Nameserver, ipv6.Prefix, ipv6.Gateway, ipv6.Nameserver),
		"-device", boardOf(machine).device("virtio-net") + ",netdev=card,mac=" + machine.Card.MAC,
	}
}
