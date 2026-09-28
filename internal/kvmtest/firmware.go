package kvmtest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/firmware"
)

// Firmware is miso's pinned UEFI firmware, its package kept in miso's own
// cache, fetched once.
func Firmware(t *testing.T) firmware.Firmware {
	t.Helper()

	found, err := firmware.Ready(t.Context(), download.Open(basesDir(t), http.DefaultClient))
	require.NoError(t, err)

	return found
}
