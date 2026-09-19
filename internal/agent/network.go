package agent

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/link"
	"github.com/The127/miso/internal/protocol"
)

// macWait and macTimeout bound the wait for the MAC of a run to be free.
const (
	macWait    = 50 * time.Millisecond
	macTimeout = 10 * time.Second
)

// addCard gives the run of a process a network card of its own on top of
// the builder's, with the address and gateway of the network, and answers
// how to remove it once the run has ended.
func addCard(pid int, network *protocol.Network) (func(), error) {
	address, err := netip.ParsePrefix(network.Address)
	if err != nil {
		return nil, fmt.Errorf("address of the run: %w", err)
	}

	if !address.Addr().Is4() {
		return nil, fmt.Errorf("address of the run: %s is not IPv4", address)
	}

	gateway, err := netip.ParseAddr(network.Gateway)
	if err != nil {
		return nil, fmt.Errorf("gateway of the run: %w", err)
	}

	if !gateway.Is4() {
		return nil, fmt.Errorf("gateway of the run: %s is not IPv4", gateway)
	}

	// the card is known by the MAC the host gave it, its name and place
	// differ between builders
	var mac []byte
	for _, part := range strings.Split(network.Card, ":") {
		octet, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return nil, fmt.Errorf("card of the run %s: %w", network.Card, err)
		}

		mac = append(mac, byte(octet))
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

	builderCards, err := builder.Ask(unix.RTM_GETLINK, unix.NLM_F_DUMP, link.CardHeader(0, 0, 0))
	if err != nil {
		return nil, fmt.Errorf("find the builder's card: %w", err)
	}

	var parent int32
	var name string
	for _, card := range builderCards {
		if bytes.Equal(link.Value(card, unix.IFLA_ADDRESS), mac) {
			parent = link.Index(card)
			name = link.Name(card)
		}
	}

	if parent == 0 {
		return nil, fmt.Errorf("the builder has no card %s", network.Card)
	}

	// the builder's card carries the runs and has no address of its own, or
	// IPv6 gives it one the moment it is up. A kernel without IPv6 has
	// nothing to switch off
	disable := "/proc/sys/net/ipv6/conf/" + name + "/disable_ipv6"
	if err := os.WriteFile(disable, []byte("1"), 0o600); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("switch off IPv6 on the builder's card: %w", err)
	}

	// nothing else in the builder brings its card up, and a card on a card
	// that is down never has a carrier
	if _, err := builder.Ask(unix.RTM_SETLINK, 0, link.CardHeader(parent, unix.IFF_UP, unix.IFF_UP)); err != nil {
		return nil, fmt.Errorf("bring up the builder's card: %w", err)
	}

	// some kernels look for the name among the builder's cards, where eth0
	// is taken, so the card arrives under a name of its own and is renamed
	arriving := fmt.Sprintf("run%d", pid)
	// the MAC follows from the address, so a run sees the same one every
	// time, and two runs on one address collide loudly instead of taking
	// turns in the gateway's table. 02 is a MAC of our own making
	local := address.Addr().As4()
	own := append([]byte{0x02, 0x00}, local[:]...)
	_, err = builder.Ask(unix.RTM_NEWLINK, unix.NLM_F_CREATE|unix.NLM_F_EXCL, link.CardHeader(0, 0, 0),
		link.Attribute(unix.IFLA_IFNAME, []byte(arriving+"\x00")),
		link.Attribute(unix.IFLA_ADDRESS, own),
		link.Attribute(unix.IFLA_LINK, binary.NativeEndian.AppendUint32(nil, uint32(parent))),
		link.Attribute(unix.IFLA_NET_NS_FD, binary.NativeEndian.AppendUint32(nil, uint32(namespace.Fd()))), //nolint:gosec // a file descriptor fits in 32 bits
		link.Attribute(unix.IFLA_LINKINFO, link.Attribute(unix.IFLA_INFO_KIND, []byte("macvlan"))),
	)
	if err != nil {
		return nil, fmt.Errorf("create card: %w", err)
	}

	run, err := link.Open(namespace)
	if err != nil {
		return nil, err
	}

	// kept open on success, it holds the run's network until the card is
	// gone. A card that failed half way is gone at once, or it holds the
	// MAC and address until the kernel removes the run's network
	kept := false
	var index int32
	defer func() {
		if kept {
			return
		}

		if index != 0 {
			_, _ = run.Ask(unix.RTM_DELLINK, 0, link.CardHeader(index, 0, 0))
		}

		_ = run.Close()
	}()

	cards, err := run.Ask(unix.RTM_GETLINK, unix.NLM_F_DUMP, link.CardHeader(0, 0, 0))
	if err != nil {
		return nil, fmt.Errorf("find card: %w", err)
	}

	for _, card := range cards {
		if link.Name(card) != arriving {
			continue
		}

		index = link.Index(card)
		if _, err := run.Ask(unix.RTM_SETLINK, 0, link.CardHeader(index, 0, 0), link.Attribute(unix.IFLA_IFNAME, []byte("eth0\x00"))); err != nil {
			return nil, fmt.Errorf("name card: %w", err)
		}

		// the address before the route, which the kernel only takes to a
		// gateway it can reach
		_, err = run.Ask(unix.RTM_NEWADDR, unix.NLM_F_CREATE|unix.NLM_F_EXCL, link.AddressHeader(index, address.Bits()),
			link.Attribute(unix.IFA_LOCAL, local[:]),
			link.Attribute(unix.IFA_ADDRESS, local[:]),
		)
		if err != nil {
			return nil, fmt.Errorf("address card %s: %w", address, err)
		}

		// a card that holds the same MAC may still be on its way out, in a
		// network the kernel removes on its own time after the run that made
		// it, so the MAC gets a while to become free
		for waited := time.Duration(0); ; waited += macWait {
			_, err := run.Ask(unix.RTM_SETLINK, 0, link.CardHeader(index, unix.IFF_UP, unix.IFF_UP))
			if err == nil {
				break
			}

			if !errors.Is(err, unix.EADDRINUSE) || waited >= macTimeout {
				return nil, fmt.Errorf("bring up card for %s: %w", address, err)
			}

			time.Sleep(macWait)
		}

		via := gateway.As4()
		_, err = run.Ask(unix.RTM_NEWROUTE, unix.NLM_F_CREATE|unix.NLM_F_EXCL, link.RouteHeader(),
			link.Attribute(unix.RTA_GATEWAY, via[:]),
			link.Attribute(unix.RTA_OIF, binary.NativeEndian.AppendUint32(nil, uint32(index))), //nolint:gosec // an index is positive
		)
		if err != nil {
			return nil, fmt.Errorf("route via %s: %w", gateway, err)
		}

		kept = true
		// the kernel removes the network of a run some time after it ended,
		// and until then its cards hold MACs and addresses a next run needs.
		// Every card goes, the run may have made more than its own
		return func() {
			cards, _ := run.Ask(unix.RTM_GETLINK, unix.NLM_F_DUMP, link.CardHeader(0, 0, 0))
			for _, card := range cards {
				if link.Flags(card)&unix.IFF_LOOPBACK != 0 {
					continue
				}

				_, _ = run.Ask(unix.RTM_DELLINK, 0, link.CardHeader(link.Index(card), 0, 0))
			}

			_ = run.Close()
		}, nil
	}

	return nil, fmt.Errorf("find card: %s is not in the run", arriving)
}
