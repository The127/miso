package runnet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net/netip"
	"slices"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/link"
)

// the parts of tc that x/sys does not name
const (
	clsact        = 0xFFFFFFF1
	egress        = 0xFFFFFFF3
	flowerAct     = 3
	flowerEthType = 8
	flowerProto   = 9
	flowerDst     = 12
	flowerDstMask = 13
	flowerDst6    = 16
	flowerDst6Msk = 17
	flowerICMPv6  = 55
	router        = 133
	solicit       = 135
	advert        = 136
	actKind       = 1
	actOptions    = 2
	gactParms     = 2
	pass          = 0
	shot          = 2
)

// fenceHost lets only ARP, IPv4 and IPv6 to the internet leave the builder's
// card, and nothing for the gateway or a loopback, which QEMU takes to the
// host's loopback.
// A run sets up its own card as it likes, so the fence sits where it cannot
// reach.
func fenceHost(builder *link.Conn, card int32, gateway netip.Addr) error {
	qdisc := link.TrafficHeader(card, 0xFFFF0000, clsact, 0)
	// the runs before this one made it already
	if _, err := builder.Ask(unix.RTM_NEWQDISC, unix.NLM_F_CREATE|unix.NLM_F_EXCL, qdisc, link.Attribute(unix.TCA_KIND, []byte("clsact\x00"))); err != nil && !errors.Is(err, unix.EEXIST) {
		return fmt.Errorf("fence the host off: %w", err)
	}

	// the first that matches decides
	filters := []struct {
		kind   uint16
		keys   []byte
		action uint32
	}{
		{unix.ETH_P_IP, destination(netip.PrefixFrom(gateway, 32)), shot},
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
		{unix.ETH_P_IPV6, destination(netip.MustParsePrefix("2000::/3")), pass},
		{unix.ETH_P_ALL, nil, shot},
	}
	for i, f := range filters {
		if err := addFilter(builder, card, uint32(i+1), f.kind, f.keys, f.action); err != nil {
			return fmt.Errorf("fence the host off: %w", err)
		}
	}

	return nil
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
		return slices.Concat(link.Attribute(flowerDst6, address), link.Attribute(flowerDst6Msk, mask))
	}

	return slices.Concat(link.Attribute(flowerDst, address), link.Attribute(flowerDstMask, mask))
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
