package builder_test

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
)

func TestAHostWithAnIPv4NameserverResolvesOverIPv4(t *testing.T) {
	// act
	resolving, err := builder.HostResolving(strings.NewReader("nameserver 127.0.0.53\n"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, builder.Resolving{IPv4: true}, resolving)
}

func TestAHostWithAnIPv6NameserverResolvesOverIPv6(t *testing.T) {
	// act
	resolving, err := builder.HostResolving(strings.NewReader("nameserver ::1\n"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, builder.Resolving{IPv6: true}, resolving)
}

func TestANameserverNamedWithAnInterfaceCountsAsItsAddress(t *testing.T) {
	// act
	resolving, err := builder.HostResolving(strings.NewReader("nameserver 192.168.1.1%eth0\n"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, builder.Resolving{IPv4: true}, resolving)
}

func TestANameserverAfterATabCounts(t *testing.T) {
	// act
	resolving, err := builder.HostResolving(strings.NewReader("nameserver\t1.1.1.1\n"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, builder.Resolving{IPv4: true}, resolving)
}

func TestWhatFollowsTheNameserversAddressIsIgnored(t *testing.T) {
	// act
	resolving, err := builder.HostResolving(strings.NewReader("nameserver 1.1.1.1 # home router\n"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, builder.Resolving{IPv4: true}, resolving)
}

func TestTheNameserversAddressEndsAtATab(t *testing.T) {
	// act
	resolving, err := builder.HostResolving(strings.NewReader("nameserver 1.1.1.1\t# home router\n"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, builder.Resolving{IPv4: true}, resolving)
}

func TestAResolvConfThatCannotBeReadFails(t *testing.T) {
	// arrange
	broken := errors.New("input/output error")

	// act
	_, err := builder.HostResolving(iotest.ErrReader(broken))

	// assert
	assert.ErrorIs(t, err, broken)
}

func TestAHostThatNamesNoNameserverResolvesInNeitherFamily(t *testing.T) {
	// act
	resolving, err := builder.HostResolving(strings.NewReader("search example.com\noptions edns0\n"))

	// assert
	require.NoError(t, err)
	// libslirp from 4.8 on falls back to 127.0.0.1 and ::1 here, before it
	// to nothing, and miso cannot tell which one QEMU runs with. A lookup
	// that fails at once beats one that waits for a timeout
	assert.Equal(t, builder.Resolving{}, resolving)
}
