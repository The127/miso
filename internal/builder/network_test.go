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
