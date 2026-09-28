package vsockns_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/vsockns"
)

// opened is a namespace held for the test, which is skipped on a host that
// cannot keep vsock private.
func opened(t *testing.T) *vsockns.Namespace {
	t.Helper()

	namespace, err := vsockns.Open()
	if errors.Is(err, vsockns.ErrNotPrivate) {
		t.Skip("this host cannot keep vsock private")
	}

	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, namespace.Close()) })

	return namespace
}
