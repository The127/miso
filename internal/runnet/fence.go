package runnet

import (
	"fmt"
	"net/netip"
	"slices"

	"github.com/florianl/go-tc"
	"golang.org/x/sys/unix"
)

// the ICMPv6 types of neighbor discovery
const (
	router  = 133
	solicit = 135
	advert  = 136
)

// fenceHost keeps a run away from the host the builder runs on and lets
// everything else through. A run sets up its own card as it likes, so the
// fence sits on the builder's card, where the run cannot reach it.
func fenceHost(card int32, wanted settings) error {
	fence, err := tc.Open(&tc.Config{})
	if err != nil {
		return fmt.Errorf("fence the host off: %w", err)
	}

	defer func() { _ = fence.Close() }()

	if err := addClsact(fence, card); err != nil {
		return fmt.Errorf("fence the host off: %w", err)
	}

	for i, what := range rules(wanted) {
		if err := addFilter(fence, card, uint16(i+1), what); err != nil {
			return fmt.Errorf("fence the host off: %w", err)
		}
	}

	return nil
}

// rules are the fence's filters in the order the kernel tries them, the
// first that matches decides. A filter names the kind of frame it is for,
// so the families never meet and only the order within one matters.
func rules(wanted settings) []rule {
	return slices.Concat(ipv4Rules(wanted.ipv4), ipv6Rules(wanted.ipv6), []rule{
		// how a run finds the gateway's MAC in IPv4
		{unix.ETH_P_ARP, tc.Flower{}, pass},
		// whatever is left is neither IP nor ARP and has no way out
		{unix.ETH_P_ALL, tc.Flower{}, shot},
	})
}

// ipv4Rules drop every IPv4 address a host takes for itself and pass the
// rest, because IPv4 has no prefix that means the internet.
func ipv4Rules(wanted family) []rule {
	return slices.Concat([]rule{
		{unix.ETH_P_IP, destination(netip.PrefixFrom(wanted.gateway, 32)), shot},
		// QEMU checks no address, so a frame a run makes itself would reach
		// the host's loopback
		{unix.ETH_P_IP, destination(netip.MustParsePrefix("127.0.0.0/8")), shot},
		// and QEMU takes the unspecified address there as well. Only the
		// address itself does, but nothing routes the rest of the range
		// anywhere either
		{unix.ETH_P_IP, destination(netip.MustParsePrefix("0.0.0.0/8")), shot},
		// and the broadcast address, which QEMU rewrites where it rewrites
		// the gateway
		{unix.ETH_P_IP, destination(netip.MustParsePrefix("255.255.255.255/32")), shot},
		// a host delivers multicast to its own listeners, so a run could
		// speak to whatever the builder runs, avahi and the like
		{unix.ETH_P_IP, destination(netip.MustParsePrefix("224.0.0.0/4")), shot},
		// nothing routes link local anywhere, and a builder in a cloud holds
		// its own credentials at 169.254.169.254
		{unix.ETH_P_IP, destination(netip.MustParsePrefix("169.254.0.0/16")), shot},
	}, nameserverRules(unix.ETH_P_IP, wanted.nameserver), []rule{
		{unix.ETH_P_IP, tc.Flower{}, pass},
	})
}

// ipv6Rules name what a run may reach instead, because two prefixes there
// cover the whole internet and every private network.
func ipv6Rules(wanted family) []rule {
	return slices.Concat([]rule{
		// how a run finds the gateway's MAC and asks for its routes in
		// IPv6, which QEMU answers itself
		{unix.ETH_P_IPV6, icmpv6(router), pass},
		{unix.ETH_P_IPV6, icmpv6(solicit), pass},
		{unix.ETH_P_IPV6, icmpv6(advert), pass},
	}, nameserverRules(unix.ETH_P_IPV6, wanted.nameserver), []rule{
		// QEMU takes its whole range to the host's loopback, the gateway and
		// the run's neighbours with it
		{unix.ETH_P_IPV6, destination(wanted.address.Masked()), shot},
		{unix.ETH_P_IPV6, destination(netip.MustParsePrefix("2000::/3")), pass},
		// a host's own network, as 192.168 and the like are in IPv4
		{unix.ETH_P_IPV6, destination(netip.MustParsePrefix("fc00::/7")), pass},
		// how a host with IPv6 alone reaches what has IPv4 alone
		{unix.ETH_P_IPV6, destination(netip.MustParsePrefix("64:ff9b::/96")), pass},
		// and the range a network picks its own such prefix from
		{unix.ETH_P_IPV6, destination(netip.MustParsePrefix("64:ff9b:1::/48")), pass},
	})
}

// nameserverRules let a query through to the nameserver and nothing else,
// and are none at all where the host named no resolver. QEMU hands the
// nameserver to the resolver of the host it runs on, and before libslirp
// 4.3.1 it did so on every port, not port 53 alone, which put the host's
// own resolver within reach of a run.
func nameserverRules(kind uint16, nameserver netip.Addr) []rule {
	if !nameserver.IsValid() {
		return nil
	}

	return []rule{
		{kind, dns(nameserver, unix.IPPROTO_UDP), pass},
		{kind, dns(nameserver, unix.IPPROTO_TCP), pass},
		{kind, destination(netip.PrefixFrom(nameserver, nameserver.BitLen())), shot},
	}
}
