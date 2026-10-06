package firmware_test

import (
	"context"
	"errors"
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
	found, err := firmware.ReadyFrom(context.Background(), store, download.Pin{URL: "https://snapshot.invalid/ovmf.deb", Digest: digest}, "amd64")

	// assert
	require.NoError(t, err)
	assert.Equal(t, firmware.Firmware{Code: []byte("the code"), Vars: []byte("the vars")}, found)
}

func TestAFirmwareForAnArchMisoHasNoFirmwareForIsRefusedNamingIt(t *testing.T) {
	// arrange
	store := download.Open(t.TempDir(), http.DefaultClient)

	// act
	_, err := firmware.Ready(context.Background(), store, "sparc")

	// assert
	assert.ErrorContains(t, err, "sparc")
}

func TestAnArm64FirmwareIsFetchedFromDebiansAAVMFPackage(t *testing.T) {
	// arrange
	var asked string
	refuse := roundTrip(func(request *http.Request) (*http.Response, error) {
		asked = request.URL.String()

		return nil, errors.New("offline")
	})
	store := download.Open(t.TempDir(), &http.Client{Transport: refuse})

	// act
	_, err := firmware.Ready(context.Background(), store, "arm64")

	// assert
	require.Error(t, err)
	assert.Contains(t, asked, "/qemu-efi-aarch64_2025.02-8%2Bdeb13u1_all.deb")
}

// roundTrip is an HTTP transport that is a function.
type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
