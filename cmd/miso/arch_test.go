package main_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/The127/miso/cmd/miso"
	"github.com/The127/miso/internal/download"
)

func TestAHostMisoDoesNotBuildOnIsRefusedNamingItAndWhereMisoBuilds(t *testing.T) {
	// act
	err := main.HostRefusal("riscv64")

	// assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "riscv64")
	assert.ErrorContains(t, err, "amd64 and arm64")
}

func TestABuildForTheBuildersOwnArchitectureFetchesNoEmulator(t *testing.T) {
	// arrange
	offline := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("offline")
	})}
	store := download.Open(t.TempDir(), offline)

	// act
	emulators, err := main.EmulatorsFor(context.Background(), store, main.BuilderArch)

	// assert
	require.NoError(t, err)
	assert.Empty(t, emulators)
}

// roundTrip is an HTTP transport that is a function.
type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
