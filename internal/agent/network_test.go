//go:build vmtest

package agent_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/layer"
	"github.com/The127/miso/internal/protocol"
)

// online is the network the VM tests hand a run, as a build hands it the
// network of the builder.
func online(t *testing.T) *protocol.Network {
	t.Helper()

	card := os.Getenv("MISO_VMTEST_MAC")
	require.NotEmpty(t, card, "MISO_VMTEST_MAC names no card")
	address := os.Getenv("MISO_VMTEST_ADDRESS")
	require.NotEmpty(t, address, "MISO_VMTEST_ADDRESS names no address")
	gateway := os.Getenv("MISO_VMTEST_GATEWAY")
	require.NotEmpty(t, gateway, "MISO_VMTEST_GATEWAY names no gateway")

	return &protocol.Network{Card: card, Address: address, Gateway: gateway}
}

func TestAnOnlineRunHasACardBesidesItsLoopback(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: "ls /sys/class/net"}
	var out bytes.Buffer

	// act
	_, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "eth0\nlo\n", out.String())
}

func TestARunReachesAServiceBeyondTheBuilder(t *testing.T) {
	// arrange
	service := os.Getenv("MISO_VMTEST_SERVICE")
	require.NotEmpty(t, service, "MISO_VMTEST_SERVICE names no service")
	host, port, _ := strings.Cut(service, ":")
	worker := mountedBase(t, t.TempDir())
	// bounded, so a run that cannot connect fails before the test does
	connect := fmt.Sprintf("timeout 5 bash -c 'exec 3<>/dev/tcp/%s/%s && cat <&3'", host, port)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: connect}
	var out bytes.Buffer

	// act
	_, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "miso\n", out.String())
}

func TestARunWithAnAddressThatIsNotIPv4FailsNamingIt(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	network := online(t)
	network.Address = "fec0::15/64"
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: network, Command: "true"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	assert.ErrorContains(t, err, "fec0::15/64")
}

func TestARunWithAGatewayThatIsNotIPv4FailsNamingIt(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	network := online(t)
	network.Gateway = "fec0::2"
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: network, Command: "true"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	assert.ErrorContains(t, err, "fec0::2")
}

func TestARunWhoseCardTheBuilderLacksFailsNamingIt(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	network := online(t)
	network.Card = "52:54:00:00:00:99"
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: network, Command: "true"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	assert.ErrorContains(t, err, "52:54:00:00:00:99")
}

func TestARunsCardHasTheSameMACEveryTime(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	first := protocol.Run{Key: "first", Layers: []string{"base"}, Network: online(t), Command: "cat /sys/class/net/eth0/address"}
	second := protocol.Run{Key: "second", Layers: []string{"base"}, Network: online(t), Command: "cat /sys/class/net/eth0/address"}
	var firstOut, secondOut bytes.Buffer
	_, err := worker.Run(context.Background(), first, &firstOut)
	require.NoError(t, err)

	// act
	_, err = worker.Run(context.Background(), second, &secondOut)

	// assert
	require.NoError(t, err)
	assert.Equal(t, firstOut.String(), secondOut.String())
}

func TestARunThatFailsToStartGivesItsAddressBack(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	work, err := layer.Open(layers).Begin("bare")
	require.NoError(t, err)
	require.NoError(t, work.Finish())
	failing := protocol.Run{Key: "failing", Layers: []string{"bare"}, Network: online(t), Command: "true"}
	_, err = worker.Run(context.Background(), failing, io.Discard)
	require.ErrorContains(t, err, "/bin/sh")
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: "true"}

	// act
	code, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
}

func TestARunWhoseNetworkFailsGivesItsAddressBack(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	unreachable := online(t)
	// outside the run's network, so the card is made and the route fails
	unreachable.Gateway = "192.0.2.1"
	failing := protocol.Run{Key: "failing", Layers: []string{"base"}, Network: unreachable, Command: "true"}
	_, err := worker.Run(context.Background(), failing, io.Discard)
	require.ErrorContains(t, err, "192.0.2.1")
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: "true"}

	// act
	code, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
}

func TestTheBuildersCardHasNoAddress(t *testing.T) {
	// arrange
	mac := os.Getenv("MISO_VMTEST_MAC")
	var card string
	addresses, err := filepath.Glob("/sys/class/net/*/address")
	require.NoError(t, err)
	for _, path := range addresses {
		address, err := os.ReadFile(path)
		require.NoError(t, err)
		if strings.TrimSpace(string(address)) == mac {
			card = filepath.Base(filepath.Dir(path))
		}
	}

	require.NotEmpty(t, card, "no card has the MAC %s", mac)
	worker := mountedBase(t, t.TempDir())
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: "true"}

	// act
	_, err = worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	ipv6, err := os.ReadFile("/proc/net/if_inet6")
	require.NoError(t, err)
	for line := range strings.Lines(string(ipv6)) {
		fields := strings.Fields(line)
		assert.NotEqual(t, card, fields[len(fields)-1], "the builder's card has %s", fields[0])
	}
}

func TestARunCleansUpTheCardsItMakes(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	// the MAC a run on 10.0.2.16 gets, which the kernel checks once a card is up
	making := protocol.Run{Key: "making", Layers: []string{"base"}, Network: online(t), Command: "ip link add m1 link eth0 address 02:00:0a:00:02:10 type macvlan && ip link set m1 up"}
	code, err := worker.Run(context.Background(), making, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	next := online(t)
	next.Address = "10.0.2.16/24"
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: next, Command: "true"}

	// act
	code, err = worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
}
