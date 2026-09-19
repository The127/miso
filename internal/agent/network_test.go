//go:build vmtest

package agent_test

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
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
	address6 := os.Getenv("MISO_VMTEST_ADDRESS6")
	require.NotEmpty(t, address6, "MISO_VMTEST_ADDRESS6 names no address")
	gateway6 := os.Getenv("MISO_VMTEST_GATEWAY6")
	require.NotEmpty(t, gateway6, "MISO_VMTEST_GATEWAY6 names no gateway")
	nameserver := os.Getenv("MISO_VMTEST_NAMESERVER")
	require.NotEmpty(t, nameserver, "MISO_VMTEST_NAMESERVER names no nameserver")
	nameserver6 := os.Getenv("MISO_VMTEST_NAMESERVER6")
	require.NotEmpty(t, nameserver6, "MISO_VMTEST_NAMESERVER6 names no nameserver")

	return &protocol.Network{
		Card: card,
		IPv4: protocol.Family{Address: address, Gateway: gateway, Nameserver: nameserver},
		IPv6: protocol.Family{Address: address6, Gateway: gateway6, Nameserver: nameserver6},
	}
}

// started starts a run and returns once the run printed its first line,
// which must be the line. What it answers waits for the end of the run and
// answers what the run printed after that line.
func started(t *testing.T, worker *agent.Agent, run protocol.Run, line string) func() string {
	t.Helper()

	reader, writer := io.Pipe()
	done := make(chan error, 1)
	go func() {
		_, err := worker.Run(context.Background(), run, writer)
		_ = writer.Close()
		done <- err
	}()

	lines := bufio.NewReader(reader)
	first, err := lines.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, line, first)
	var rest bytes.Buffer
	copied := make(chan struct{})
	go func() {
		_, _ = io.Copy(&rest, lines)
		close(copied)
	}()

	return func() string {
		t.Helper()

		require.NoError(t, <-done)
		<-copied

		return rest.String()
	}
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

func TestAnOnlineRunHasItsIPv6AddressAtOnce(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	network := online(t)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: network, Command: "ip -6 -o addr show dev eth0 scope global"}
	var out bytes.Buffer

	// act
	_, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "inet6 "+network.IPv6.Address)
	assert.NotContains(t, out.String(), "tentative")
}

func TestAnOnlineRunHasAnIPv6RouteThroughItsGateway(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	network := online(t)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: network, Command: "ip -6 route show default"}
	var out bytes.Buffer

	// act
	_, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "default via "+network.IPv6.Gateway+" dev eth0")
}

func TestARunsIPv6RoutesAreOnlyTheOnesTheHostGave(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	// the kernel asks QEMU for routes a moment after the card is up
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: "sleep 3; ip -6 route show default"}
	var out bytes.Buffer

	// act
	_, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	// a second gateway may show as a second route or as a second way of one
	assert.Equal(t, 1, strings.Count(out.String(), "via"), out.String())
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

func TestARunReachesAServiceBeyondTheBuilderOverIPv6(t *testing.T) {
	// arrange
	service := os.Getenv("MISO_VMTEST_SERVICE6")
	require.NotEmpty(t, service, "MISO_VMTEST_SERVICE6 names no service")
	host, port, err := net.SplitHostPort(service)
	require.NoError(t, err)
	worker := mountedBase(t, t.TempDir())
	// bounded, so a run that cannot connect fails before the test does
	connect := fmt.Sprintf("timeout 5 bash -c 'exec 3<>/dev/tcp/%s/%s && cat <&3'", host, port)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: connect}
	var out bytes.Buffer

	// act
	_, err = worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "miso\n", out.String())
}

func TestARunReachesAServiceOnAUniqueLocalAddress(t *testing.T) {
	// arrange
	service := os.Getenv("MISO_VMTEST_LOCAL6")
	require.NotEmpty(t, service, "MISO_VMTEST_LOCAL6 names no service")
	host, port, err := net.SplitHostPort(service)
	require.NoError(t, err)
	worker := mountedBase(t, t.TempDir())
	// bounded, so a run that cannot connect fails before the test does
	connect := fmt.Sprintf("timeout 5 bash -c 'exec 3<>/dev/tcp/%s/%s && cat <&3'", host, port)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: connect}
	var out bytes.Buffer

	// act
	_, err = worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "miso\n", out.String())
}

func TestARunReachesAnIPv4ServiceThroughNAT64(t *testing.T) {
	// arrange
	service := os.Getenv("MISO_VMTEST_MAPPED6")
	require.NotEmpty(t, service, "MISO_VMTEST_MAPPED6 names no service")
	host, port, err := net.SplitHostPort(service)
	require.NoError(t, err)
	worker := mountedBase(t, t.TempDir())
	// bounded, so a run that cannot connect fails before the test does
	connect := fmt.Sprintf("timeout 5 bash -c 'exec 3<>/dev/tcp/%s/%s && cat <&3'", host, port)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: connect}
	var out bytes.Buffer

	// act
	_, err = worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "miso\n", out.String())
}

func TestARunWithAnAddressThatIsNotIPv4FailsNamingIt(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	network := online(t)
	network.IPv4.Address = "fec0::15/64"
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
	network.IPv4.Gateway = "fec0::2"
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: network, Command: "true"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	assert.ErrorContains(t, err, "fec0::2")
}

func TestARunWithAnIPv6AddressThatIsNotIPv6FailsNamingIt(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	network := online(t)
	network.IPv6.Address = "10.0.2.99/24"
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: network, Command: "true"}

	// act
	_, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	assert.ErrorContains(t, err, "10.0.2.99/24")
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
	unreachable.IPv4.Gateway = "192.0.2.1"
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
	next.IPv4.Address = "10.0.2.16/24"
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: next, Command: "true"}

	// act
	code, err = worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
}

func TestTwoRunsTalkAtTheSameTime(t *testing.T) {
	// arrange
	meet := os.Getenv("MISO_VMTEST_MEET")
	require.NotEmpty(t, meet, "MISO_VMTEST_MEET names no meeting point")
	host, port, _ := strings.Cut(meet, ":")
	worker := mountedBase(t, t.TempDir())
	connect := fmt.Sprintf("timeout 15 bash -c 'exec 3<>/dev/tcp/%s/%s && cat <&3'", host, port)
	var outs [2]bytes.Buffer
	var errs [2]error
	var wait sync.WaitGroup

	// act
	for i, address := range []string{"10.0.2.15/24", "10.0.2.16/24"} {
		network := online(t)
		network.IPv4.Address = address
		run := protocol.Run{Key: fmt.Sprintf("run%d", i), Layers: []string{"base"}, Network: network, Command: connect}
		wait.Go(func() { _, errs[i] = worker.Run(context.Background(), run, &outs[i]) })
	}

	wait.Wait()

	// assert
	for i := range outs {
		require.NoError(t, errs[i])
		assert.Equal(t, "met\n", outs[i].String())
	}
}

func TestACardARunHidesInANetworkOfItsOwnIsGoneForTheNextRun(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	// the MAC a run on 10.0.2.16 gets, on a card the agent cannot reach. The
	// kernel removes that network soon after the run, and this test guards
	// that it stays soon enough
	hiding := strings.Join([]string{
		"set -e",
		"unshare -n sleep 1000 &",
		"inner=$!",
		"sleep 0.3",
		"ip link add m1 link eth0 address 02:00:0a:00:02:10 type macvlan",
		"ip link set m1 netns $inner",
		"nsenter -t $inner -n ip link set m1 up",
	}, "\n")
	making := protocol.Run{Key: "making", Layers: []string{"base"}, Network: online(t), Command: hiding}
	code, err := worker.Run(context.Background(), making, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	next := online(t)
	next.IPv4.Address = "10.0.2.16/24"
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: next, Command: "true"}

	// act
	code, err = worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
}

func TestARunWaitsUntilItsMACIsFree(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	// the MAC a run on 10.0.2.16 gets, held a while longer
	holding := protocol.Run{
		Key: "holding", Layers: []string{"base"}, Network: online(t),
		Command: "ip link add m1 link eth0 address 02:00:0a:00:02:10 type macvlan && ip link set m1 up && echo holding && sleep 2",
	}
	finish := started(t, worker, holding, "holding\n")
	next := online(t)
	next.IPv4.Address = "10.0.2.16/24"
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: next, Command: "true"}

	// act
	code, err := worker.Run(context.Background(), run, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	finish()
}

func TestARunCannotReachAnotherRun(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	listening := protocol.Run{
		Key: "listening", Layers: []string{"base"}, Network: online(t),
		Command: `perl -MIO::Socket::INET -e '$| = 1; alarm 4; $s = IO::Socket::INET->new(LocalPort => 9, Listen => 1, ReuseAddr => 1) or die $!; print "listening\n"; $s->accept and print "accepted\n"'`,
	}
	heard := started(t, worker, listening, "listening\n")
	other := online(t)
	other.IPv4.Address = "10.0.2.16/24"
	connecting := protocol.Run{
		Key: "connecting", Layers: []string{"base"}, Network: other,
		Command: "timeout 3 bash -c 'exec 3<>/dev/tcp/10.0.2.15/9 && echo reached'",
	}
	var out bytes.Buffer

	// act
	_, err := worker.Run(context.Background(), connecting, &out)

	// assert
	require.NoError(t, err)
	assert.NotContains(t, out.String(), "reached")
	assert.NotContains(t, heard(), "accepted")
}

func TestARunCannotReachTheHostsLoopback(t *testing.T) {
	// arrange
	loopback := os.Getenv("MISO_VMTEST_HOST")
	require.NotEmpty(t, loopback, "MISO_VMTEST_HOST names no service")
	host, port, _ := strings.Cut(loopback, ":")
	worker := mountedBase(t, t.TempDir())
	connect := fmt.Sprintf("timeout 3 bash -c 'exec 3<>/dev/tcp/%s/%s && cat <&3'", host, port)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: connect}
	var out bytes.Buffer

	// act
	_, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.NotContains(t, out.String(), "loopback")
}

func TestARunThatSetsUpIPv6ItselfCannotReachTheHostsLoopback(t *testing.T) {
	// arrange
	loopback := os.Getenv("MISO_VMTEST_HOST")
	require.NotEmpty(t, loopback, "MISO_VMTEST_HOST names no service")
	_, port, _ := strings.Cut(loopback, ":")
	address := os.Getenv("MISO_VMTEST_ADDRESS6")
	require.NotEmpty(t, address, "MISO_VMTEST_ADDRESS6 names no address")
	gateway := os.Getenv("MISO_VMTEST_GATEWAY6")
	require.NotEmpty(t, gateway, "MISO_VMTEST_GATEWAY6 names no gateway")
	worker := mountedBase(t, t.TempDir())
	// a run may configure its own card, and QEMU maps its IPv6 gateway to
	// the host's loopback as it does in IPv4
	connect := strings.Join([]string{
		"set -e",
		// replace, the agent or QEMU may have set them already
		fmt.Sprintf("ip -6 addr replace %s dev eth0 nodad", address),
		fmt.Sprintf("ip -6 route replace default via %s", gateway),
		fmt.Sprintf("timeout 3 bash -c 'exec 3<>/dev/tcp/%s/%s && cat <&3' || true", gateway, port),
	}, "\n")
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: connect}
	var out bytes.Buffer

	// act
	code, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code, out.String())
	assert.NotContains(t, out.String(), "loopback")
}

//go:embed testdata/query.py
var query string

func TestARunResolvesANameThroughTheNameserverOverIPv6(t *testing.T) {
	// arrange
	nameserver := os.Getenv("MISO_VMTEST_NAMESERVER6")
	require.NotEmpty(t, nameserver, "MISO_VMTEST_NAMESERVER6 names no nameserver")
	worker := mountedBase(t, t.TempDir())
	asking := fmt.Sprintf("python3 - %s miso.test AAAA <<'EOF'\n%sEOF", nameserver, query)
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: online(t), Command: asking}
	var out bytes.Buffer

	// act
	code, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code, out.String())
	assert.Equal(t, "2001:db8::53\n", out.String())
}

//go:embed testdata/syn.py
var syn string

func TestARunCannotReachTheHostsLoopbackWithAFrameOfItsOwn(t *testing.T) {
	// arrange
	loopback := os.Getenv("MISO_VMTEST_HOST")
	require.NotEmpty(t, loopback, "MISO_VMTEST_HOST names no service")
	_, port, _ := strings.Cut(loopback, ":")
	network := online(t)
	address, _, _ := strings.Cut(network.IPv4.Address, "/")
	worker := mountedBase(t, t.TempDir())
	// the attempt is dropped, but it teaches the run the gateway's MAC
	sending := strings.Join([]string{
		"set -e",
		fmt.Sprintf("timeout 1 bash -c 'exec 3<>/dev/tcp/%s/1' || true", network.IPv4.Gateway),
		fmt.Sprintf("gateway=$(ip neigh show %s | awk '{print $5}')", network.IPv4.Gateway),
		fmt.Sprintf("python3 - \"$gateway\" %s 127.0.0.1 %s <<'EOF'\n%sEOF", address, port, syn),
	}, "\n")
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: network, Command: sending}
	var out bytes.Buffer

	// act
	code, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code, out.String())
	assert.NotContains(t, out.String(), "reached")
}

func TestARunCannotReachTheHostsLoopbackThroughTheUnspecifiedAddress(t *testing.T) {
	// arrange
	loopback := os.Getenv("MISO_VMTEST_HOST")
	require.NotEmpty(t, loopback, "MISO_VMTEST_HOST names no service")
	_, port, _ := strings.Cut(loopback, ":")
	network := online(t)
	address, _, _ := strings.Cut(network.IPv4.Address, "/")
	worker := mountedBase(t, t.TempDir())
	// a run's own kernel turns a connect to it into one to the run's
	// loopback, so only a frame a run makes itself gets out
	sending := strings.Join([]string{
		"set -e",
		fmt.Sprintf("timeout 1 bash -c 'exec 3<>/dev/tcp/%s/1' || true", network.IPv4.Gateway),
		fmt.Sprintf("gateway=$(ip neigh show %s | awk '{print $5}')", network.IPv4.Gateway),
		fmt.Sprintf("python3 - \"$gateway\" %s 0.0.0.0 %s <<'EOF'\n%sEOF", address, port, syn),
	}, "\n")
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Network: network, Command: sending}
	var out bytes.Buffer

	// act
	code, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	require.Equal(t, 0, code, out.String())
	assert.NotContains(t, out.String(), "reached")
}

func TestAnOfflineRunHasOnlyItsLoopback(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	run := protocol.Run{Key: "run", Layers: []string{"base"}, Command: "ls /sys/class/net"}
	var out bytes.Buffer

	// act
	_, err := worker.Run(context.Background(), run, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "lo\n", out.String())
}
