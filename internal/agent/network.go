package agent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"slices"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/link"
	"github.com/The127/miso/internal/protocol"
)

// addCard gives the run of a process a network card of its own on top of
// the builder's, with the address and gateway of the network, and answers
// how to remove it once the run has ended.
func addCard(pid int, network *protocol.Network) (func(), error) {
	wanted, err := readSettings(network)
	if err != nil {
		return nil, err
	}

	namespace, err := os.Open(fmt.Sprintf("/proc/%d/ns/net", pid))
	if err != nil {
		return nil, err
	}

	defer func() { _ = namespace.Close() }()

	builder, err := link.Open(nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = builder.Close() }()

	parent, err := builderCard(builder, wanted.card, network.Card)
	if err != nil {
		return nil, err
	}

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

	qdisc := binary.NativeEndian.AppendUint32([]byte{unix.AF_UNSPEC, 0, 0, 0}, uint32(parent)) //nolint:gosec // an index is positive
	qdisc = binary.NativeEndian.AppendUint32(qdisc, 0xFFFF0000)
	qdisc = binary.NativeEndian.AppendUint32(qdisc, clsact)
	qdisc = binary.NativeEndian.AppendUint32(qdisc, 0)
	// the runs before this one made it already
	if _, err := builder.Ask(unix.RTM_NEWQDISC, unix.NLM_F_CREATE|unix.NLM_F_EXCL, qdisc, link.Attribute(unix.TCA_KIND, []byte("clsact\x00"))); err != nil && !errors.Is(err, unix.EEXIST) {
		return nil, fmt.Errorf("fence the host off: %w", err)
	}

	ipv4 := binary.BigEndian.AppendUint16(nil, unix.ETH_P_IP)
	filter := binary.NativeEndian.AppendUint32([]byte{unix.AF_UNSPEC, 0, 0, 0}, uint32(parent)) //nolint:gosec // an index is positive
	filter = binary.NativeEndian.AppendUint32(filter, 1)
	filter = binary.NativeEndian.AppendUint32(filter, egress)
	filter = binary.NativeEndian.AppendUint32(filter, 1<<16|uint32(binary.NativeEndian.Uint16(ipv4)))
	gateway := wanted.gateway.As4()
	drop := make([]byte, 20)
	binary.NativeEndian.PutUint32(drop[8:], shot)
	action := link.Attribute(1|unix.NLA_F_NESTED, slices.Concat(
		link.Attribute(actKind, []byte("gact\x00")),
		link.Attribute(actOptions|unix.NLA_F_NESTED, link.Attribute(gactParms, drop)),
	))
	options := slices.Concat(
		link.Attribute(flowerEthType, ipv4),
		link.Attribute(flowerDst, gateway[:]),
		link.Attribute(flowerDstMask, []byte{0xFF, 0xFF, 0xFF, 0xFF}),
		link.Attribute(flowerAct|unix.NLA_F_NESTED, action),
	)
	if _, err := builder.Ask(unix.RTM_NEWTFILTER, unix.NLM_F_CREATE|unix.NLM_F_REPLACE, filter,
		link.Attribute(unix.TCA_KIND, []byte("flower\x00")), link.Attribute(unix.TCA_OPTIONS|unix.NLA_F_NESTED, options)); err != nil {
		return nil, fmt.Errorf("fence the host off: %w", err)
	}

	// some kernels look for the name among the builder's cards, where eth0
	// is taken, so the card arrives under a name of its own and is renamed
	arriving := fmt.Sprintf("run%d", pid)
	if err := createCard(builder, parent, namespace, arriving, wanted.mac()); err != nil {
		return nil, err
	}

	// open until the cards are gone, it holds the run's network
	run, err := link.Open(namespace)
	if err != nil {
		return nil, err
	}

	index, err := configureCard(run, arriving, wanted)
	if err != nil {
		// a card that failed half way goes at once, or it holds the MAC and
		// address until the kernel removes the run's network
		if index != 0 {
			_, _ = run.Ask(unix.RTM_DELLINK, 0, link.CardHeader(index, 0, 0))
		}

		_ = run.Close()

		return nil, err
	}

	return func() {
		removeCards(run)
		_ = run.Close()
	}, nil
}
