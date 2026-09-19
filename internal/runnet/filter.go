package runnet

import (
	"encoding/binary"
	"errors"
	"net/netip"
	"slices"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/link"
)

// the parts of tc that x/sys does not name
const (
	clsact         = 0xFFFFFFF1
	egress         = 0xFFFFFFF3
	flowerAct      = 3
	flowerEthType  = 8
	flowerProto    = 9
	flowerDst      = 12
	flowerDstMask  = 13
	flowerDst6     = 16
	flowerDst6Mask = 17
	flowerTCPDst   = 19
	flowerUDPDst   = 21
	flowerICMPv6   = 55
	actKind        = 1
	actOptions     = 2
	gactParms      = 2
	pass           = 0
	shot           = 2
)

// addClsact gives a card the queue that filters hang on, unless it has it.
func addClsact(builder *link.Conn, card int32) error {
	qdisc := link.TrafficHeader(card, 0xFFFF0000, clsact, 0)
	// the runs before this one made it already
	_, err := builder.Ask(unix.RTM_NEWQDISC, unix.NLM_F_CREATE|unix.NLM_F_EXCL, qdisc, link.Attribute(unix.TCA_KIND, []byte("clsact\x00")))
	if errors.Is(err, unix.EEXIST) {
		return nil
	}

	return err
}

// destination is the key of a filter that matches a prefix of IPv4 or
// IPv6.
func destination(prefix netip.Prefix) []byte {
	address := prefix.Addr().AsSlice()
	mask := make([]byte, len(address))
	for i := range prefix.Bits() {
		mask[i/8] |= 0x80 >> (i % 8)
	}

	if prefix.Addr().Is6() {
		return slices.Concat(link.Attribute(flowerDst6, address), link.Attribute(flowerDst6Mask, mask))
	}

	return slices.Concat(link.Attribute(flowerDst, address), link.Attribute(flowerDstMask, mask))
}

// dns is the key of a filter that matches a query to a nameserver, which
// is what slirp hands to the resolver of the host it runs on. flower reads
// a port only where the protocol names one, so each protocol is its own
// filter.
func dns(nameserver netip.Addr, over byte) []byte {
	attribute := uint16(flowerUDPDst)
	if over == unix.IPPROTO_TCP {
		attribute = flowerTCPDst
	}

	return slices.Concat(
		destination(netip.PrefixFrom(nameserver, nameserver.BitLen())),
		link.Attribute(flowerProto, []byte{over}),
		link.Attribute(attribute, binary.BigEndian.AppendUint16(nil, 53)),
	)
}

// icmpv6 is the key of a filter that matches ICMPv6 of a type.
func icmpv6(kind byte) []byte {
	return slices.Concat(link.Attribute(flowerProto, []byte{unix.IPPROTO_ICMPV6}), link.Attribute(flowerICMPv6, []byte{kind}))
}

// addFilter sets the filter of a priority on the egress of a card, which
// does an action to the frames of a kind that match the keys.
func addFilter(builder *link.Conn, card int32, priority uint32, kind uint16, keys []byte, action uint32) error {
	ethType := binary.BigEndian.AppendUint16(nil, kind)
	filter := link.TrafficHeader(card, 1, egress, priority<<16|uint32(binary.NativeEndian.Uint16(ethType)))
	parms := make([]byte, 20)
	binary.NativeEndian.PutUint32(parms[8:], action)
	actions := link.Attribute(1|unix.NLA_F_NESTED, slices.Concat(
		link.Attribute(actKind, []byte("gact\x00")),
		link.Attribute(actOptions|unix.NLA_F_NESTED, link.Attribute(gactParms, parms)),
	))
	// a key on every kind but all, which has no type to match
	if kind != unix.ETH_P_ALL {
		keys = slices.Concat(link.Attribute(flowerEthType, ethType), keys)
	}

	options := slices.Concat(keys, link.Attribute(flowerAct|unix.NLA_F_NESTED, actions))
	_, err := builder.Ask(unix.RTM_NEWTFILTER, unix.NLM_F_CREATE|unix.NLM_F_REPLACE, filter,
		link.Attribute(unix.TCA_KIND, []byte("flower\x00")), link.Attribute(unix.TCA_OPTIONS|unix.NLA_F_NESTED, options))

	return err
}
