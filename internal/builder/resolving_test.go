package builder_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/builder"
)

func TestAHostWithAnIPv4NameserverResolvesOverIPv4(t *testing.T) {
	// act
	resolving := builder.HostResolving(strings.NewReader("nameserver 127.0.0.53\n"))

	// assert
	assert.Equal(t, builder.Resolving{IPv4: true}, resolving)
}

func TestAHostWithAnIPv6NameserverResolvesOverIPv6(t *testing.T) {
	// act
	resolving := builder.HostResolving(strings.NewReader("nameserver ::1\n"))

	// assert
	assert.Equal(t, builder.Resolving{IPv6: true}, resolving)
}

func TestANameserverNamedWithAnInterfaceCountsAsItsAddress(t *testing.T) {
	// act
	resolving := builder.HostResolving(strings.NewReader("nameserver 192.168.1.1%eth0\n"))

	// assert
	assert.Equal(t, builder.Resolving{IPv4: true}, resolving)
}
