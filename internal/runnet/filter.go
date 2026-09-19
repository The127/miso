package runnet

import (
	"net"
	"net/netip"

	"github.com/florianl/go-tc"
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
