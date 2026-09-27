package builder_test

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
)

func TestTheRunsLookForTheMACOfTheBuildersCard(t *testing.T) {
	// act
	card, network := builder.Network(resolvingBoth)

	// assert
	require.NotEmpty(t, card.MAC)
	assert.Equal(t, card.MAC, network.Card)
}

func TestARunGetsAnIPv4AddressOfItsOwnOnTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network(resolvingBoth)

	// assert
	prefix := parsedPrefix(t, card.IPv4.Prefix)
	address := parsedPrefix(t, network.IPv4.Address)
	assert.True(t, address.Addr().Is4())
	assert.True(t, prefix.Contains(address.Addr()))
	assert.Equal(t, prefix.Bits(), address.Bits())
	assert.NotEqual(t, card.IPv4.Gateway, address.Addr().String())
	assert.NotEqual(t, card.IPv4.Nameserver, address.Addr().String())
}

func TestARunGoesOutThroughTheIPv4GatewayOfTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network(resolvingBoth)

	// assert
	prefix := parsedPrefix(t, card.IPv4.Prefix)
	gateway := parsedAddr(t, card.IPv4.Gateway)
	assert.True(t, prefix.Contains(gateway))
	assert.Equal(t, card.IPv4.Gateway, network.IPv4.Gateway)
}

func TestARunResolvesThroughTheIPv4NameserverOfTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network(resolvingBoth)

	// assert
	prefix := parsedPrefix(t, card.IPv4.Prefix)
	nameserver := parsedAddr(t, card.IPv4.Nameserver)
	assert.True(t, prefix.Contains(nameserver))
	assert.Equal(t, card.IPv4.Nameserver, network.IPv4.Nameserver)
}

func TestARunGetsAnIPv6AddressOfItsOwnOnTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network(resolvingBoth)

	// assert
	prefix := parsedPrefix(t, card.IPv6.Prefix)
	address := parsedPrefix(t, network.IPv6.Address)
	assert.True(t, address.Addr().Is6())
	assert.True(t, prefix.Contains(address.Addr()))
	assert.Equal(t, prefix.Bits(), address.Bits())
	assert.NotEqual(t, card.IPv6.Gateway, address.Addr().String())
	assert.NotEqual(t, card.IPv6.Nameserver, address.Addr().String())
}

func TestARunGoesOutThroughTheIPv6GatewayOfTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network(resolvingBoth)

	// assert
	prefix := parsedPrefix(t, card.IPv6.Prefix)
	gateway := parsedAddr(t, card.IPv6.Gateway)
	assert.True(t, prefix.Contains(gateway))
	assert.Equal(t, card.IPv6.Gateway, network.IPv6.Gateway)
}

func TestARunResolvesThroughTheIPv6NameserverOfTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network(resolvingBoth)

	// assert
	prefix := parsedPrefix(t, card.IPv6.Prefix)
	nameserver := parsedAddr(t, card.IPv6.Nameserver)
	assert.True(t, prefix.Contains(nameserver))
	assert.Equal(t, card.IPv6.Nameserver, network.IPv6.Nameserver)
}

func TestARunOnAHostThatResolvesOnlyOverIPv4GetsNoIPv6Nameserver(t *testing.T) {
	// act
	card, network := builder.Network(builder.Resolving{IPv4: true})

	// assert
	assert.NotEmpty(t, card.IPv6.Nameserver)
	assert.Empty(t, network.IPv6.Nameserver)
	assert.Equal(t, card.IPv4.Nameserver, network.IPv4.Nameserver)
}

func TestARunOnAHostThatResolvesOnlyOverIPv6GetsNoIPv4Nameserver(t *testing.T) {
	// act
	card, network := builder.Network(builder.Resolving{IPv6: true})

	// assert
	assert.NotEmpty(t, card.IPv4.Nameserver)
	assert.Empty(t, network.IPv4.Nameserver)
	assert.Equal(t, card.IPv6.Nameserver, network.IPv6.Nameserver)
}

// resolvingBoth is a host that resolves names in both families.
var resolvingBoth = builder.Resolving{IPv4: true, IPv6: true}

func parsedPrefix(t *testing.T, text string) netip.Prefix {
	t.Helper()

	parsed, err := netip.ParsePrefix(text)
	require.NoError(t, err)

	return parsed
}

func parsedAddr(t *testing.T, text string) netip.Addr {
	t.Helper()

	parsed, err := netip.ParseAddr(text)
	require.NoError(t, err)

	return parsed
}
