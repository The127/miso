package check_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/check"
)

func TestABootWithoutAVsockNamespaceIsRefused(t *testing.T) {
	// arrange
	boot := check.Boot{Dir: t.TempDir(), Patience: time.Second}

	// act
	_, err := boot.Run(t.Context(), []string{"true"})

	// assert
	assert.ErrorIs(t, err, check.ErrNoNamespace)
}
