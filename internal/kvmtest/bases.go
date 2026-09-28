package kvmtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// basesDir is where miso keeps what it fetches, so the tests fetch each
// thing once, as a build would.
func basesDir(t *testing.T) string {
	t.Helper()

	cache, err := os.UserCacheDir()
	require.NoError(t, err)

	return filepath.Join(cache, "miso", "bases")
}
