package runnet

import (
	"errors"
	"net"
	"net/netip"

	"github.com/florianl/go-tc"
	"github.com/florianl/go-tc/core"
	"golang.org/x/sys/unix"
)

// the actions of gact that a fence uses
const (
	pass = 0
	shot = 2
)

// egress is where the filters on what leaves a card hang, the other half of
// the clsact queue.
const egress = tc.HandleIngress + 2

// rule is one filter of a fence: what it does to the frames of a kind that
// match its keys.
type rule struct {
	kind   uint16
	keys   tc.Flower
	action uint32
}

// addClsact gives a card the queue that filters hang on, unless it has it.
func addClsact(fence *tc.Tc, card int32) error {
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
		return err
	}

	return nil
}

// addFilter sets a rule at a priority on the egress of a card.
func addFilter(fence *tc.Tc, card int32, priority uint16, what rule) error {
	keys := what.keys
	// every kind but all, which has no type to match
	if what.kind != unix.ETH_P_ALL {
		kind := what.kind
		keys.KeyEthType = &kind
	}

	actions := []*tc.Action{{Kind: "gact", Gact: &tc.Gact{Parms: &tc.GactParms{Action: what.action}}}}
	keys.Actions = &actions
	filter := tc.Object{
		Msg: tc.Msg{
			Family:  unix.AF_UNSPEC,
			Ifindex: uint32(card), //nolint:gosec // an index is never negative
			// named, so a later run replaces this filter instead of the
			// kernel giving it a handle of its own and keeping both
			Handle: 1,
			Parent: egress,
			Info:   core.FilterInfo(priority, what.kind),
		},
		Attribute: tc.Attribute{Kind: "flower", Flower: &keys},
	}

	// replaced, not added, because the run before this one left its own
	// filters on the card
	return fence.Filter().Replace(&filter)
}

// destination matches a prefix of IPv4 or IPv6.
func destination(prefix netip.Prefix) tc.Flower {
	address := net.IP(prefix.Addr().AsSlice())
	mask := net.IP(net.CIDRMask(prefix.Bits(), prefix.Addr().BitLen()))
	if prefix.Addr().Is6() {
		return tc.Flower{KeyIPv6Dst: &address, KeyIPv6DstMask: &mask}
	}

	return tc.Flower{KeyIPv4Dst: &address, KeyIPv4DstMask: &mask}
}

// dns matches a query to a nameserver, which is what slirp hands to the
// resolver of the host it runs on. flower reads a port only where the
// protocol names one, so each protocol is its own filter.
func dns(nameserver netip.Addr, over uint8) tc.Flower {
	keys := destination(netip.PrefixFrom(nameserver, nameserver.BitLen()))
	port := uint16(53)
	keys.KeyIPProto = &over
	if over == unix.IPPROTO_TCP {
		keys.KeyTCPDst = &port
	} else {
		keys.KeyUDPDst = &port
	}

	return keys
}

// icmpv6 matches ICMPv6 of a type.
func icmpv6(kind uint8) tc.Flower {
	protocol := uint8(unix.IPPROTO_ICMPV6)

	return tc.Flower{KeyIPProto: &protocol, KeyIcmpv6Type: &kind}
}
