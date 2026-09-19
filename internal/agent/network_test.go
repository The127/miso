//go:build vmtest

package agent_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
