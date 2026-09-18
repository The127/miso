package protocol_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/protocol"
)

func TestTheSerialOfABaseIsTheStartOfItsHash(t *testing.T) {
	// arrange
	digest := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	// act
	serial := protocol.Serial(digest)

	// assert
	assert.Equal(t, "0123456789abcdef0123", serial)
}
