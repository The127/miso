package builder

import (
	"testing"
	"time"
)

// Patience has Ask wait for the agent to listen for d instead, for the rest
// of a test.
func Patience(t *testing.T, d time.Duration) {
	t.Helper()

	was := patience
	patience = d

	t.Cleanup(func() { patience = was })
}
