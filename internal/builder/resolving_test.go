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
