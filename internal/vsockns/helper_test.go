package vsockns_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/vsockns"
)

func TestAQuestionTheHelperDoesNotKnowIsAnsweredAsFailed(t *testing.T) {
	// arrange
	namespace, err := vsockns.Open()
	if errors.Is(err, vsockns.ErrNotPrivate) {
		t.Skip("this host cannot keep vsock private")
	}

	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, namespace.Close()) })

	// act
	_, err = vsockns.Ask(namespace, "shutdown")

	// assert
	assert.ErrorContains(t, err, "unknown question shutdown")
}
