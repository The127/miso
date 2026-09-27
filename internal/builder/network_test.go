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
	card, network := builder.Network()

	// assert
	require.NotEmpty(t, card.MAC)
	assert.Equal(t, card.MAC, network.Card)
}

func TestARunGetsAnIPv4AddressOfItsOwnOnTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network()

	// assert
	prefix, err := netip.ParsePrefix(card.IPv4.Prefix)
	require.NoError(t, err)
	address, err := netip.ParsePrefix(network.IPv4.Address)
	require.NoError(t, err)
	assert.True(t, address.Addr().Is4())
	assert.True(t, prefix.Contains(address.Addr()))
	assert.Equal(t, prefix.Bits(), address.Bits())
	assert.NotEqual(t, card.IPv4.Gateway, address.Addr().String())
	assert.NotEqual(t, card.IPv4.Nameserver, address.Addr().String())
}

func TestARunGoesOutThroughTheIPv4GatewayOfTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network()

	// assert
	prefix, err := netip.ParsePrefix(card.IPv4.Prefix)
	require.NoError(t, err)
	gateway, err := netip.ParseAddr(card.IPv4.Gateway)
	require.NoError(t, err)
	assert.True(t, prefix.Contains(gateway))
	assert.Equal(t, card.IPv4.Gateway, network.IPv4.Gateway)
}

func TestARunResolvesThroughTheIPv4NameserverOfTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network()

	// assert
	prefix, err := netip.ParsePrefix(card.IPv4.Prefix)
	require.NoError(t, err)
	nameserver, err := netip.ParseAddr(card.IPv4.Nameserver)
	require.NoError(t, err)
	assert.True(t, prefix.Contains(nameserver))
	assert.Equal(t, card.IPv4.Nameserver, network.IPv4.Nameserver)
}

func TestARunGetsAnIPv6AddressOfItsOwnOnTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network()

	// assert
	prefix, err := netip.ParsePrefix(card.IPv6.Prefix)
	require.NoError(t, err)
	address, err := netip.ParsePrefix(network.IPv6.Address)
	require.NoError(t, err)
	assert.True(t, address.Addr().Is6())
	assert.True(t, prefix.Contains(address.Addr()))
	assert.Equal(t, prefix.Bits(), address.Bits())
	assert.NotEqual(t, card.IPv6.Gateway, address.Addr().String())
	assert.NotEqual(t, card.IPv6.Nameserver, address.Addr().String())
}

func TestARunGoesOutThroughTheIPv6GatewayOfTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network()

	// assert
	prefix, err := netip.ParsePrefix(card.IPv6.Prefix)
	require.NoError(t, err)
	gateway, err := netip.ParseAddr(card.IPv6.Gateway)
	require.NoError(t, err)
	assert.True(t, prefix.Contains(gateway))
	assert.Equal(t, card.IPv6.Gateway, network.IPv6.Gateway)
}

func TestARunResolvesThroughTheIPv6NameserverOfTheCardsNetwork(t *testing.T) {
	// act
	card, network := builder.Network()

	// assert
	prefix, err := netip.ParsePrefix(card.IPv6.Prefix)
	require.NoError(t, err)
	nameserver, err := netip.ParseAddr(card.IPv6.Nameserver)
	require.NoError(t, err)
	assert.True(t, prefix.Contains(nameserver))
	assert.Equal(t, card.IPv6.Nameserver, network.IPv6.Nameserver)
}
