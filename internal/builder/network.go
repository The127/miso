package builder

import (
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
)

// mac is the builder's card's, 52:54:00 as QEMU's own and then mis.
const mac = "52:54:00:6d:69:73"

// Resolving is which address families the host resolves names in. QEMU
// forwards a query only to a resolver of the query's own family.
type Resolving struct {
	IPv4 bool
	IPv6 bool
}

// Network is the builder VM's card on QEMU's user network and the network a
// run gets through it, from one set of numbers so the two agree.
func Network(resolving Resolving) (qemu.Card, protocol.Network) {
	ipv4 := qemu.Family{Prefix: "10.0.2.0/24", Gateway: "10.0.2.2", Nameserver: "10.0.2.3"}
	ipv6 := qemu.Family{Prefix: "fd6d:6973:6f00::/64", Gateway: "fd6d:6973:6f00::2", Nameserver: "fd6d:6973:6f00::3"}

	card := qemu.Card{MAC: mac, IPv4: ipv4, IPv6: ipv6}
	network := protocol.Network{
		Card: mac,
		IPv4: run(ipv4, "10.0.2.15/24", resolving.IPv4),
		IPv6: run(ipv6, "fd6d:6973:6f00::15/64", resolving.IPv6),
	}

	return card, network
}

// run is a run's side of one family of the card's network, with the address
// given. It gets the card's nameserver only where the host resolves, one
// that forwards nowhere would cost every lookup a timeout.
func run(family qemu.Family, address string, resolves bool) protocol.Family {
	run := protocol.Family{Address: address, Gateway: family.Gateway}
	if resolves {
		run.Nameserver = family.Nameserver
	}

	return run
}
