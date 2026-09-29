package vsockns_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/vsockns/vsocknstest"
)

func TestASocketOfTheNamespaceIsNotInMisosOwn(t *testing.T) {
	// arrange
	namespace := vsocknstest.Private(t)

	// act
	socket, err := namespace.Socket()

	// assert
	require.NoError(t, err)
	t.Cleanup(func() { _ = socket.Close() })
	assert.NotEqual(t, cookie(t, ours(t)), cookie(t, socket))
}
