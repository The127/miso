package runnet

import (
	"errors"
	"fmt"
	"net/netip"

	"github.com/florianl/go-tc"
	"github.com/florianl/go-tc/core"
	"golang.org/x/sys/unix"
)

// the ICMPv6 types of neighbor discovery
const (
	router  = 133
	solicit = 135
	advert  = 136
)

// fenceHost lets only ARP, IPv4 and IPv6 to the internet or a host's own
// network leave the builder's card, and nothing for the gateway or a
// loopback, which QEMU takes to the host's loopback. A run sets up its own
// card as it likes, so the fence sits where it cannot reach.
func fenceHost(card int32, wanted settings) error {
	fence, err := tc.Open(&tc.Config{})
	if err != nil {
		return fmt.Errorf("fence the host off: %w", err)
	}

	defer func() { _ = fence.Close() }()

	queue := tc.Object{
		Msg: tc.Msg{
			Family:  unix.AF_UNSPEC,
			Ifindex: uint32(card), //nolint:gosec // an index is never negative
			Handle:  core.BuildHandle(tc.HandleRoot, 0),
			Parent:  tc.HandleIngress,
		},
		Attribute: tc.Attribute{Kind: "clsact"},
	}
	// the runs before this one made it already
	if err := fence.Qdisc().Add(&queue); err != nil && !errors.Is(err, unix.EEXIST) {
		return fmt.Errorf("fence the host off: %w", err)
	}

	// the first that matches decides
	type rule struct {
		kind   uint16
		keys   tc.Flower
		action uint32
	}

	filters := []rule{
		{unix.ETH_P_IP, destination(netip.PrefixFrom(wanted.ipv4.gateway, 32)), shot},
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
	}

	// QEMU hands the nameserver to the resolver of the host it runs on, and
	// before libslirp 4.3.1 it did so on every port, not port 53 alone,
	// which put the host's own resolver within reach of a run
	if wanted.ipv4.nameserver.IsValid() {
		filters = append(filters,
			rule{unix.ETH_P_IP, dns(wanted.ipv4.nameserver, unix.IPPROTO_UDP), pass},
			rule{unix.ETH_P_IP, dns(wanted.ipv4.nameserver, unix.IPPROTO_TCP), pass},
			rule{unix.ETH_P_IP, destination(netip.PrefixFrom(wanted.ipv4.nameserver, 32)), shot},
		)
	}

	filters = append(filters, []rule{
		{unix.ETH_P_IP, tc.Flower{}, pass},
		{unix.ETH_P_ARP, tc.Flower{}, pass},
		// how a run finds the gateway's MAC and asks for its routes in
		// IPv6, which QEMU answers itself
		{unix.ETH_P_IPV6, icmpv6(router), pass},
		{unix.ETH_P_IPV6, icmpv6(solicit), pass},
		{unix.ETH_P_IPV6, icmpv6(advert), pass},
	}...)

	// the nameserver sits in the range the next rule drops, so a query has
	// to pass before it
	if wanted.ipv6.nameserver.IsValid() {
		filters = append(filters,
			rule{unix.ETH_P_IPV6, dns(wanted.ipv6.nameserver, unix.IPPROTO_UDP), pass},
			rule{unix.ETH_P_IPV6, dns(wanted.ipv6.nameserver, unix.IPPROTO_TCP), pass},
		)
	}

	filters = append(filters, []rule{
		// QEMU takes its whole range to the host's loopback, the gateway and
		// the run's neighbours with it
		{unix.ETH_P_IPV6, destination(wanted.ipv6.address.Masked()), shot},
		{unix.ETH_P_IPV6, destination(netip.MustParsePrefix("2000::/3")), pass},
		// a host's own network, as 192.168 and the like are in IPv4
		{unix.ETH_P_IPV6, destination(netip.MustParsePrefix("fc00::/7")), pass},
		// how a host with IPv6 alone reaches what has IPv4 alone
		{unix.ETH_P_IPV6, destination(netip.MustParsePrefix("64:ff9b::/96")), pass},
		{unix.ETH_P_ALL, tc.Flower{}, shot},
	}...)

	for i, f := range filters {
		keys := f.keys
		if f.kind != unix.ETH_P_ALL {
			kind := f.kind
			keys.KeyEthType = &kind
		}

		actions := []*tc.Action{{Kind: "gact", Gact: &tc.Gact{Parms: &tc.GactParms{Action: f.action}}}}
		keys.Actions = &actions
		filter := tc.Object{
			Msg: tc.Msg{
				Family:  unix.AF_UNSPEC,
				Ifindex: uint32(card), //nolint:gosec // an index is never negative
				// named, so a later run replaces this filter instead of the
				// kernel giving it a handle of its own and keeping both
				Handle: 1,
				Parent: egress,
				Info:   core.FilterInfo(uint16(i+1), f.kind),
			},
			Attribute: tc.Attribute{Kind: "flower", Flower: &keys},
		}
		// replaced, not added, because the run before this one left its own
		// filters on the card
		if err := fence.Filter().Replace(&filter); err != nil {
			return fmt.Errorf("fence the host off: %w", err)
		}
	}

	return nil
}
