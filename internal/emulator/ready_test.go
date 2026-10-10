package emulator_test

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
	"github.com/The127/miso/internal/emulator"
)

func TestTheEmulatorOfAnArm64ImageIsTakenFromThePinnedPackage(t *testing.T) {
	// arrange
	pkg := debtest.Package(t, map[string]string{
		"./usr/bin/qemu-aarch64": "the emulator",
		"./usr/bin/qemu-x86_64":  "another emulator",
	})
	dir := t.TempDir()
	digest := downloadtest.Kept(t, dir, pkg)
	store := download.Open(dir, http.DefaultClient)

	// act
	found, err := emulator.ReadyFrom(context.Background(), store, download.Pin{URL: "https://snapshot.invalid/qemu-user.deb", Digest: digest}, "arm64")

	// assert
	require.NoError(t, err)
	assert.Equal(t, []byte("the emulator"), found)
}

func TestAnArm64BuilderRunsAnAmd64ImageWithQEMUOfItsOwnArchitecture(t *testing.T) {
	// arrange
	var asked string
	refuse := roundTrip(func(request *http.Request) (*http.Response, error) {
		asked = request.URL.String()

		return nil, errors.New("offline")
	})
	store := download.Open(t.TempDir(), &http.Client{Transport: refuse})

	// act
	_, err := emulator.Ready(context.Background(), store, "arm64", "amd64")

	// assert
	require.Error(t, err)
	assert.Contains(t, asked, "/qemu-user_10.0.13%2Bds-0%2Bdeb13u1_arm64.deb")
}

func TestAnEmulatorForABuilderMisoHasNoQEMUForIsRefusedNamingIt(t *testing.T) {
	// arrange
	store := download.Open(t.TempDir(), http.DefaultClient)

	// act
	_, err := emulator.Ready(context.Background(), store, "sparc", "amd64")

	// assert
	assert.ErrorContains(t, err, "sparc")
}

func TestAnEmulatorForAnImageMisoHasNoQEMUForIsRefusedNamingItBeforeAnyFetch(t *testing.T) {
	// arrange
	var asked string
	refuse := roundTrip(func(request *http.Request) (*http.Response, error) {
		asked = request.URL.String()

		return nil, errors.New("offline")
	})
	store := download.Open(t.TempDir(), &http.Client{Transport: refuse})

	// act
	_, err := emulator.Ready(context.Background(), store, "amd64", "sparc")

	// assert
	assert.ErrorContains(t, err, "sparc")
	assert.Empty(t, asked)
}

// roundTrip is an HTTP transport that is a function.
type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
