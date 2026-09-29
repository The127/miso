package check_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/vsockns"
)

// private is a vsock namespace held for the test, which is skipped on a host
// that cannot keep vsock private.
func private(t *testing.T) *vsockns.Namespace {
	t.Helper()

	namespace, err := vsockns.Open()
	if errors.Is(err, vsockns.ErrNotPrivate) {
		t.Skip("this host cannot keep vsock private")
	}

	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, namespace.Close()) })

	return namespace
}

func TestABootWithoutAVsockNamespaceIsRefused(t *testing.T) {
	// arrange
	boot := check.Boot{Dir: t.TempDir(), Patience: time.Second}

	// act
	_, err := boot.Run(t.Context(), []string{"true"})

	// assert
	assert.ErrorIs(t, err, check.ErrNoNamespace)
}
