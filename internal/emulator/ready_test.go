package emulator_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"path/filepath"
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
	found, err := emulator.ReadyFrom(context.Background(), store, download.Pin{URL: "https://snapshot.invalid/qemu-user.deb", Digest: digest}, "arm64", digestOf("the emulator"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, []byte("the emulator"), found)
}

func TestAnEmulatorTheStoreKeepsIsReadWithoutThePackage(t *testing.T) {
	// arrange
	dir := t.TempDir()
	digest := downloadtest.Kept(t, dir, []byte("the emulator"))
	refuse := roundTrip(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("offline")
	})
	store := download.Open(dir, &http.Client{Transport: refuse})

	// act
	found, err := emulator.ReadyFrom(context.Background(), store, download.Pin{URL: "https://snapshot.invalid/qemu-user.deb", Digest: "sha256:notkept"}, "arm64", digest)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []byte("the emulator"), found)
}

func TestAnEmulatorTakenFromThePackageIsKeptUnderItsDigest(t *testing.T) {
	// arrange
	pkg := debtest.Package(t, map[string]string{"./usr/bin/qemu-aarch64": "the emulator"})
	dir := t.TempDir()
	digest := downloadtest.Kept(t, dir, pkg)
	store := download.Open(dir, http.DefaultClient)
	program := digestOf("the emulator")

	// act
	_, err := emulator.ReadyFrom(context.Background(), store, download.Pin{URL: "https://snapshot.invalid/qemu-user.deb", Digest: digest}, "arm64", program)

	// assert
	require.NoError(t, err)
	kept, err := os.ReadFile(store.Path(program))
	require.NoError(t, err)
	assert.Equal(t, []byte("the emulator"), kept)
}

func TestAnEmulatorOtherThanItsPinIsRefusedAndNotKept(t *testing.T) {
	// arrange
	pkg := debtest.Package(t, map[string]string{"./usr/bin/qemu-aarch64": "another emulator"})
	dir := t.TempDir()
	digest := downloadtest.Kept(t, dir, pkg)
	store := download.Open(dir, http.DefaultClient)
	pinned := digestOf("the emulator")

	// act
	_, err := emulator.ReadyFrom(context.Background(), store, download.Pin{URL: "https://snapshot.invalid/qemu-user.deb", Digest: digest}, "arm64", pinned)

	// assert
	require.ErrorContains(t, err, pinned)
	assert.NoFileExists(t, store.Path(digestOf("another emulator")))
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

func TestAnArm64BuilderReadsTheAmd64EmulatorItKeptWithoutThePackage(t *testing.T) {
	// arrange
	dir := t.TempDir()
	refuse := roundTrip(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("offline")
	})
	store := download.Open(dir, &http.Client{Transport: refuse})
	kept := store.Path("sha256:9af86352fd463ad9ec25bdd6a4b891ac6b6a7f3d46a7db3da6f23e2c09094a59")
	require.NoError(t, os.MkdirAll(filepath.Dir(kept), 0o750))
	require.NoError(t, os.WriteFile(kept, []byte("the emulator"), 0o600))

	// act
	found, err := emulator.Ready(context.Background(), store, "arm64", "amd64")

	// assert
	require.NoError(t, err)
	assert.Equal(t, []byte("the emulator"), found)
}

func TestAnAmd64BuilderReadsTheArm64EmulatorItKeptWithoutThePackage(t *testing.T) {
	// arrange
	dir := t.TempDir()
	refuse := roundTrip(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("offline")
	})
	store := download.Open(dir, &http.Client{Transport: refuse})
	kept := store.Path("sha256:73cc2584f119f9e74c85c0a8549dbb6fd104ae3897a06a7500797f9e5b4d0943")
	require.NoError(t, os.MkdirAll(filepath.Dir(kept), 0o750))
	require.NoError(t, os.WriteFile(kept, []byte("the emulator"), 0o600))

	// act
	found, err := emulator.Ready(context.Background(), store, "amd64", "arm64")

	// assert
	require.NoError(t, err)
	assert.Equal(t, []byte("the emulator"), found)
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

func digestOf(content string) string {
	sum := sha256.Sum256([]byte(content))

	return "sha256:" + hex.EncodeToString(sum[:])
}
