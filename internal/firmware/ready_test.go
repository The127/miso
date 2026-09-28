package firmware_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/deb/debtest"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/download/downloadtest"
	"github.com/The127/miso/internal/firmware"
)

func TestTheFirmwareIsReadiedFromThePinnedPackage(t *testing.T) {
	// arrange
	pkg := debtest.Package(t, map[string]string{
		"./usr/share/OVMF/OVMF_CODE_4M.fd": "the code",
		"./usr/share/OVMF/OVMF_VARS_4M.fd": "the vars",
	})
	dir := t.TempDir()
	digest := downloadtest.Kept(t, dir, pkg)
	store := download.Open(dir, http.DefaultClient)

	// act
	found, err := firmware.ReadyFrom(context.Background(), store, download.Pin{URL: "https://snapshot.invalid/ovmf.deb", Digest: digest})

	// assert
	require.NoError(t, err)
	assert.Equal(t, firmware.Firmware{Code: []byte("the code"), Vars: []byte("the vars")}, found)
}
