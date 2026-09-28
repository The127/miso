package kvmtest

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/firmware"
)

// Firmware is miso's pinned UEFI firmware, its package kept in miso's own
// cache, fetched once.
func Firmware(t *testing.T) firmware.Firmware {
	t.Helper()

	cache, err := os.UserCacheDir()
	require.NoError(t, err)
	found, err := firmware.Ready(t.Context(), download.Open(filepath.Join(cache, "miso", "bases"), http.DefaultClient))
	require.NoError(t, err)

	return found
}
