package runnet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net/netip"
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
func configureCard(run *link.Conn, namespace *os.File, arriving string, wanted settings) (int32, error) {
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

		// QEMU hands out routes and addresses of its own when asked, and on a
		// timer, and a run has only what the host gave. No netlink request
		// sets it, so a thread in the run's network writes the setting
		err = link.Within(namespace, func() error {
			return os.WriteFile("/proc/sys/net/ipv6/conf/eth0/accept_ra", []byte("0"), 0o600)
		})
		if err != nil {
			return index, fmt.Errorf("refuse QEMU's routes: %w", err)
		}

		// the addresses before the routes, which the kernel only takes to a
		// gateway it can reach. The host hands out every address once, so
		// the run need not wait a second for the kernel to make sure of one
		if err := addAddress(run, index, wanted.ipv4.address, 0); err != nil {
			return index, err
		}

		if err := addAddress(run, index, wanted.ipv6.address, unix.IFA_F_NODAD); err != nil {
			return index, err
		}

		if err := bringUp(run, index, wanted.ipv4.address); err != nil {
			return index, err
		}

		for _, f := range []family{wanted.ipv4, wanted.ipv6} {
			if err := addRoute(run, index, f.gateway); err != nil {
				return index, err
			}
		}

		return index, nil
	}

	return 0, fmt.Errorf("find card: %s is not in the run", arriving)
}

// addAddress gives the card at an index an address, IFA_F_NODAD and the
// like among the flags.
func addAddress(run *link.Conn, index int32, address netip.Prefix, flags byte) error {
	local := address.Addr().AsSlice()
	_, err := run.Ask(unix.RTM_NEWADDR, unix.NLM_F_CREATE|unix.NLM_F_EXCL, link.AddressHeader(familyOf(address.Addr()), index, address.Bits(), flags),
		link.Attribute(unix.IFA_LOCAL, local),
		link.Attribute(unix.IFA_ADDRESS, local),
	)
	if err != nil {
		return fmt.Errorf("address card %s: %w", address, err)
	}

	return nil
}

// bringUp brings up the card at an index, which holds the MAC that follows
// from an address.
func bringUp(run *link.Conn, index int32, address netip.Prefix) error {
	// a card that holds the same MAC may still be on its way out, in a
	// network the kernel removes on its own time after the run that made
	// it, so the MAC gets a while to become free
	for waited := time.Duration(0); ; waited += macWait {
		_, err := run.Ask(unix.RTM_SETLINK, 0, link.CardHeader(index, unix.IFF_UP, unix.IFF_UP))
		if err == nil {
			return nil
		}

		if !errors.Is(err, unix.EADDRINUSE) || waited >= macTimeout {
			return fmt.Errorf("bring up card for %s: %w", address, err)
		}

		time.Sleep(macWait)
	}
}

// addRoute gives the card at an index the default route through a gateway.
func addRoute(run *link.Conn, index int32, gateway netip.Addr) error {
	_, err := run.Ask(unix.RTM_NEWROUTE, unix.NLM_F_CREATE|unix.NLM_F_EXCL, link.RouteHeader(familyOf(gateway)),
		link.Attribute(unix.RTA_GATEWAY, gateway.AsSlice()),
		link.Attribute(unix.RTA_OIF, binary.NativeEndian.AppendUint32(nil, uint32(index))), //nolint:gosec // an index is positive
	)
	if err != nil {
		return fmt.Errorf("route via %s: %w", gateway, err)
	}

	return nil
}

// familyOf is the address family of an address, as netlink names it.
func familyOf(address netip.Addr) byte {
	if address.Is4() {
		return unix.AF_INET
	}

	return unix.AF_INET6
}
