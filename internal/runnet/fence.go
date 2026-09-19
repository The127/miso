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
	flowerDst     = 12
	flowerDstMask = 13
	actKind       = 1
	actOptions    = 2
	gactParms     = 2
	shot          = 2
)

// fenceHost drops what leaves the builder's card for the gateway, which
// QEMU maps to the host's loopback.
func fenceHost(builder *link.Conn, card int32, gateway netip.Addr) error {
	qdisc := link.TrafficHeader(card, 0xFFFF0000, clsact, 0)
	// the runs before this one made it already
	if _, err := builder.Ask(unix.RTM_NEWQDISC, unix.NLM_F_CREATE|unix.NLM_F_EXCL, qdisc, link.Attribute(unix.TCA_KIND, []byte("clsact\x00"))); err != nil && !errors.Is(err, unix.EEXIST) {
		return fmt.Errorf("fence the host off: %w", err)
	}

	ipv4 := binary.BigEndian.AppendUint16(nil, unix.ETH_P_IP)
	filter := link.TrafficHeader(card, 1, egress, 1<<16|uint32(binary.NativeEndian.Uint16(ipv4)))
	address := gateway.As4()
	drop := make([]byte, 20)
	binary.NativeEndian.PutUint32(drop[8:], shot)
	action := link.Attribute(1|unix.NLA_F_NESTED, slices.Concat(
		link.Attribute(actKind, []byte("gact\x00")),
		link.Attribute(actOptions|unix.NLA_F_NESTED, link.Attribute(gactParms, drop)),
	))
	options := slices.Concat(
		link.Attribute(flowerEthType, ipv4),
		link.Attribute(flowerDst, address[:]),
		link.Attribute(flowerDstMask, []byte{0xFF, 0xFF, 0xFF, 0xFF}),
		link.Attribute(flowerAct|unix.NLA_F_NESTED, action),
	)
	if _, err := builder.Ask(unix.RTM_NEWTFILTER, unix.NLM_F_CREATE|unix.NLM_F_REPLACE, filter,
		link.Attribute(unix.TCA_KIND, []byte("flower\x00")), link.Attribute(unix.TCA_OPTIONS|unix.NLA_F_NESTED, options)); err != nil {
		return fmt.Errorf("fence the host off: %w", err)
	}

	return nil
}
