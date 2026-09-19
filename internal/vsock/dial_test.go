//go:build vmtest

package vsock_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestADialedConnectionIsNotInheritedByAProcess(t *testing.T) {
	// act
	_, dialed := dialing(t, 1027)

	// assert
	assert.True(t, closedOnExec(t, dialed))
}
