package runnet

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"github.com/The127/miso/internal/protocol"
)

// settings are the network the host handed a run, read.
type settings struct {
	// the MAC of the builder's card, known by it because its name and place
	// differ between builders
	card []byte
	ipv4 family
	ipv6 family
}

// family is the part of a run's network in one address family, read.
type family struct {
	address netip.Prefix
	gateway netip.Addr

	// not valid when the host named none in this family
	nameserver netip.Addr
}

// readSettings reads the network the host handed a run.
func readSettings(network *protocol.Network) (settings, error) {
	ipv4, err := readFamily(network.IPv4, "IPv4", netip.Addr.Is4)
	if err != nil {
		return settings{}, err
	}

	// an IPv4 address would pass as IPv4-mapped, which the kernel takes
	ipv6, err := readFamily(network.IPv6, "IPv6", netip.Addr.Is6)
	if err != nil {
		return settings{}, err
	}

	var card []byte
	for _, part := range strings.Split(network.Card, ":") {
		octet, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return settings{}, fmt.Errorf("card of the run %s: %w", network.Card, err)
		}

		card = append(card, byte(octet))
	}

	return settings{card: card, ipv4: ipv4, ipv6: ipv6}, nil
}

// readFamily reads one family of a run's network. The two read alike, so
// the name of the family goes in front of whatever went wrong, and own
// says which addresses belong to it.
func readFamily(handed protocol.Family, name string, own func(netip.Addr) bool) (family, error) {
	address, err := netip.ParsePrefix(handed.Address)
	if err != nil {
		return family{}, fmt.Errorf("%s address of the run: %w", name, err)
	}

	if !own(address.Addr()) {
		return family{}, fmt.Errorf("%s address of the run: %s is not %s", name, address, name)
	}

	gateway, err := netip.ParseAddr(handed.Gateway)
	if err != nil {
		return family{}, fmt.Errorf("%s gateway of the run: %w", name, err)
	}

	if !own(gateway) {
		return family{}, fmt.Errorf("%s gateway of the run: %s is not %s", name, gateway, name)
	}

	// a host without a resolver in this family names none, because slirp
	// forwards a query only to a resolver of the query's own family, so one
	// handed over anyway would cost every lookup a timeout
	var nameserver netip.Addr
	if handed.Nameserver != "" {
		nameserver, err = netip.ParseAddr(handed.Nameserver)
		if err != nil {
			return family{}, fmt.Errorf("%s nameserver of the run: %w", name, err)
		}

		if !own(nameserver) {
			return family{}, fmt.Errorf("%s nameserver of the run: %s is not %s", name, nameserver, name)
		}
	}

	return family{address: address, gateway: gateway, nameserver: nameserver}, nil
}

// mac is the MAC of the run's own card. It follows from the address, so a
// run sees the same one every time, and two runs on one address collide
// loudly instead of taking turns in the gateway's table. 02 is a MAC of our
// own making.
func (s settings) mac() []byte {
	local := s.ipv4.address.Addr().As4()

	return append([]byte{0x02, 0x00}, local[:]...)
}
