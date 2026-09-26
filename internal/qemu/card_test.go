package qemu_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachinesCardHasTheMACItAsksForOnQEMUsUserNetwork(t *testing.T) {
	// arrange
	machine := qemu.Machine{Card: &qemu.Card{MAC: "52:54:00:6d:69:73"}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	options := strings.Split(valueOf(t, args, "-netdev"), ",")
	assert.Equal(t, "user", options[0])
	assert.Contains(t, options, "id=card")
	assert.Equal(t, "virtio-net-pci,netdev=card,mac=52:54:00:6d:69:73", valueOf(t, args, "-device"))
}

func TestACardsIPv4NetworkHasTheLayoutItAsksFor(t *testing.T) {
	// arrange
	ipv4 := qemu.Family{Prefix: "10.0.2.0/24", Gateway: "10.0.2.2", Nameserver: "10.0.2.3"}
	machine := qemu.Machine{Card: &qemu.Card{MAC: "52:54:00:6d:69:73", IPv4: ipv4}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	options := strings.Split(valueOf(t, args, "-netdev"), ",")
	assert.Contains(t, options, "net=10.0.2.0/24")
	assert.Contains(t, options, "host=10.0.2.2")
	assert.Contains(t, options, "dns=10.0.2.3")
}

func TestACardsIPv6NetworkHasTheLayoutItAsksFor(t *testing.T) {
	// arrange
	ipv6 := qemu.Family{Prefix: "fd6d:6973:6f00::/64", Gateway: "fd6d:6973:6f00::2", Nameserver: "fd6d:6973:6f00::3"}
	machine := qemu.Machine{Card: &qemu.Card{MAC: "52:54:00:6d:69:73", IPv6: ipv6}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	options := strings.Split(valueOf(t, args, "-netdev"), ",")
	assert.Contains(t, options, "ipv6-net=fd6d:6973:6f00::/64")
	assert.Contains(t, options, "ipv6-host=fd6d:6973:6f00::2")
	assert.Contains(t, options, "ipv6-dns=fd6d:6973:6f00::3")
}

func TestAMachineWithNoCardHasNoNetwork(t *testing.T) {
	// arrange
	machine := qemu.Machine{}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "none", valueOf(t, args, "-nic"))
	assert.NotContains(t, args, "-netdev")
}
