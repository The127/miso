package builderkernel_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/download/downloadtest"
)

func TestTheBuilderKernelIsReadiedFromThePinnedPackage(t *testing.T) {
	// arrange
	deb := wholePackage(t, map[string]string{
		"./lib/modules/" + testRelease + "/kernel/fs/btrfs/btrfs.ko.xz": packedModule(t, "name=btrfs"),
	})
	dir := t.TempDir()
	digest := downloadtest.Kept(t, dir, deb)
	store := download.Open(dir, http.DefaultClient)

	// act
	booting, err := builderkernel.ReadyFrom(context.Background(), store, download.Pin{URL: "https://snapshot.invalid/linux-image.deb", Digest: digest}, "btrfs")

	// assert
	require.NoError(t, err)
	assert.Equal(t, testRelease, booting.Release)
	assert.Equal(t, "the kernel", string(booting.Image))
	require.Len(t, booting.Modules, 1)
	assert.Equal(t, "btrfs", booting.Modules[0].Name)
}

func TestABuilderKernelForAnArchMisoHasNoKernelForIsRefusedNamingIt(t *testing.T) {
	// arrange
	store := download.Open(t.TempDir(), http.DefaultClient)

	// act
	_, err := builderkernel.Ready(context.Background(), store, "sparc")

	// assert
	assert.ErrorContains(t, err, "sparc")
}

func TestAnArm64BuilderKernelIsFetchedFromDebiansArm64CloudKernelPackage(t *testing.T) {
	// arrange
	var asked string
	refuse := roundTrip(func(request *http.Request) (*http.Response, error) {
		asked = request.URL.String()

		return nil, errors.New("offline")
	})
	store := download.Open(t.TempDir(), &http.Client{Transport: refuse})

	// act
	_, err := builderkernel.Ready(context.Background(), store, "arm64")

	// assert
	require.Error(t, err)
	assert.Contains(t, asked, "/linux-image-6.12.107+deb13-cloud-arm64_")
}

// roundTrip is an HTTP transport that is a function.
type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
