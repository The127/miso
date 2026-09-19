package runnet

import (
	"fmt"
	"net/netip"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/link"
)

// the ICMPv6 types of neighbor discovery that x/sys does not name
const (
	router  = 133
	solicit = 135
	advert  = 136
)

// fenceHost lets only ARP, IPv4 and IPv6 to the internet or a host's own
// network leave the builder's card, and nothing for the gateway or a
// loopback, which QEMU takes to the host's loopback. A run sets up its own
// card as it likes, so the fence sits where it cannot reach.
func fenceHost(builder *link.Conn, card int32, wanted settings) error {
	if err := addClsact(builder, card); err != nil {
		return fmt.Errorf("fence the host off: %w", err)
	}

	// the first that matches decides
	filters := []struct {
		kind   uint16
		keys   []byte
		action uint32
	}{
		{unix.ETH_P_IP, destination(netip.PrefixFrom(wanted.ipv4.gateway, 32)), shot},
		// QEMU checks no address, so a frame a run makes itself would reach
		// the host's loopback
		{unix.ETH_P_IP, destination(netip.MustParsePrefix("127.0.0.0/8")), shot},
		{unix.ETH_P_IP, nil, pass},
		{unix.ETH_P_ARP, nil, pass},
		// how a run finds the gateway's MAC and asks for its routes in
		// IPv6, which QEMU answers itself
		{unix.ETH_P_IPV6, icmpv6(router), pass},
		{unix.ETH_P_IPV6, icmpv6(solicit), pass},
		{unix.ETH_P_IPV6, icmpv6(advert), pass},
		// QEMU takes its whole range to the host's loopback, the gateway and
		// the run's neighbours with it
		{unix.ETH_P_IPV6, destination(wanted.ipv6.address.Masked()), shot},
		{unix.ETH_P_IPV6, destination(netip.MustParsePrefix("2000::/3")), pass},
		// a host's own network, as 192.168 and the like are in IPv4
		{unix.ETH_P_IPV6, destination(netip.MustParsePrefix("fc00::/7")), pass},
		{unix.ETH_P_ALL, nil, shot},
	}
	for i, f := range filters {
		if err := addFilter(builder, card, uint32(i+1), f.kind, f.keys, f.action); err != nil {
			return fmt.Errorf("fence the host off: %w", err)
		}
	}

	return nil
}
