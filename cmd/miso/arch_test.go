package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/The127/miso/cmd/miso"
)

func TestAHostMisoDoesNotBuildOnIsRefusedNamingItAndWhereMisoBuilds(t *testing.T) {
	// act
	err := main.HostRefusal("riscv64")

	// assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "riscv64")
	assert.ErrorContains(t, err, "amd64 and arm64")
}
