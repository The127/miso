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
}

// readSettings reads the network the host handed a run.
func readSettings(network *protocol.Network) (settings, error) {
	address, err := netip.ParsePrefix(network.IPv4.Address)
	if err != nil {
		return settings{}, fmt.Errorf("address of the run: %w", err)
	}

	if !address.Addr().Is4() {
		return settings{}, fmt.Errorf("address of the run: %s is not IPv4", address)
	}

	gateway, err := netip.ParseAddr(network.IPv4.Gateway)
	if err != nil {
		return settings{}, fmt.Errorf("gateway of the run: %w", err)
	}

	if !gateway.Is4() {
		return settings{}, fmt.Errorf("gateway of the run: %s is not IPv4", gateway)
	}

	address6, err := netip.ParsePrefix(network.IPv6.Address)
	if err != nil {
		return settings{}, fmt.Errorf("IPv6 address of the run: %w", err)
	}

	// an IPv4 one would pass as IPv4-mapped, which the kernel takes
	if !address6.Addr().Is6() {
		return settings{}, fmt.Errorf("IPv6 address of the run: %s is not IPv6", address6)
	}

	gateway6, err := netip.ParseAddr(network.IPv6.Gateway)
	if err != nil {
		return settings{}, fmt.Errorf("IPv6 gateway of the run: %w", err)
	}

	var card []byte
	for _, part := range strings.Split(network.Card, ":") {
		octet, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return settings{}, fmt.Errorf("card of the run %s: %w", network.Card, err)
		}

		card = append(card, byte(octet))
	}

	return settings{
		card: card,
		ipv4: family{address: address, gateway: gateway},
		ipv6: family{address: address6, gateway: gateway6},
	}, nil
}

// mac is the MAC of the run's own card. It follows from the address, so a
// run sees the same one every time, and two runs on one address collide
// loudly instead of taking turns in the gateway's table. 02 is a MAC of our
// own making.
func (s settings) mac() []byte {
	local := s.ipv4.address.Addr().As4()

	return append([]byte{0x02, 0x00}, local[:]...)
}
