package agent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/link"
)

// macWait and macTimeout bound the wait for the MAC of a run to be free.
const (
	macWait    = 50 * time.Millisecond
	macTimeout = 10 * time.Second
)

// createCard makes a card with a MAC on top of the builder's card at the
// index parent, in the network namespace of a run, under a name.
func createCard(builder *link.Conn, parent int32, namespace *os.File, name string, mac []byte) error {
	_, err := builder.Ask(unix.RTM_NEWLINK, unix.NLM_F_CREATE|unix.NLM_F_EXCL, link.CardHeader(0, 0, 0),
		link.Attribute(unix.IFLA_IFNAME, []byte(name+"\x00")),
		link.Attribute(unix.IFLA_ADDRESS, mac),
		link.Attribute(unix.IFLA_LINK, binary.NativeEndian.AppendUint32(nil, uint32(parent))),              //nolint:gosec // an index is positive
		link.Attribute(unix.IFLA_NET_NS_FD, binary.NativeEndian.AppendUint32(nil, uint32(namespace.Fd()))), //nolint:gosec // a file descriptor fits in 32 bits
		link.Attribute(unix.IFLA_LINKINFO, link.Attribute(unix.IFLA_INFO_KIND, []byte("macvlan"))),
	)
	if err != nil {
		return fmt.Errorf("create card: %w", err)
	}

	return nil
}

// configureCard makes the card that arrived under a name in a run the
// run's eth0, with the address and the route of the settings. It answers
// the card's index, also when it failed half way, so the card can go.
func configureCard(run *link.Conn, arriving string, wanted settings) (int32, error) {
	cards, err := run.Ask(unix.RTM_GETLINK, unix.NLM_F_DUMP, link.CardHeader(0, 0, 0))
	if err != nil {
		return 0, fmt.Errorf("find card: %w", err)
	}

	for _, card := range cards {
		if link.Name(card) != arriving {
			continue
		}

		index := link.Index(card)
		if _, err := run.Ask(unix.RTM_SETLINK, 0, link.CardHeader(index, 0, 0), link.Attribute(unix.IFLA_IFNAME, []byte("eth0\x00"))); err != nil {
			return index, fmt.Errorf("name card: %w", err)
		}

		// the address before the route, which the kernel only takes to a
		// gateway it can reach
		local := wanted.address.Addr().As4()
		_, err = run.Ask(unix.RTM_NEWADDR, unix.NLM_F_CREATE|unix.NLM_F_EXCL, link.AddressHeader(index, wanted.address.Bits()),
			link.Attribute(unix.IFA_LOCAL, local[:]),
			link.Attribute(unix.IFA_ADDRESS, local[:]),
		)
		if err != nil {
			return index, fmt.Errorf("address card %s: %w", wanted.address, err)
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
				return index, fmt.Errorf("bring up card for %s: %w", wanted.address, err)
			}

			time.Sleep(macWait)
		}

		via := wanted.gateway.As4()
		_, err = run.Ask(unix.RTM_NEWROUTE, unix.NLM_F_CREATE|unix.NLM_F_EXCL, link.RouteHeader(),
			link.Attribute(unix.RTA_GATEWAY, via[:]),
			link.Attribute(unix.RTA_OIF, binary.NativeEndian.AppendUint32(nil, uint32(index))), //nolint:gosec // an index is positive
		)
		if err != nil {
			return index, fmt.Errorf("route via %s: %w", wanted.gateway, err)
		}

		return index, nil
	}

	return 0, fmt.Errorf("find card: %s is not in the run", arriving)
}
