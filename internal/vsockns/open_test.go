package vsockns_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/vsockns"
)

func TestANamespaceForVMsKeepsItsVsockToItselfOrIsRefused(t *testing.T) {
	// act
	namespace, err := vsockns.Open()

	// assert
	if errors.Is(err, vsockns.ErrNotPrivate) {
		return
	}

	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, namespace.Close()) })
	mode, err := namespace.Mode()
	require.NoError(t, err)
	assert.Equal(t, "local", mode)
}
