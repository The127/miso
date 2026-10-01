//go:build kvm

package main_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPruneAsksTheBuilderAndSaysWhatItRemoved(t *testing.T) {
	// arrange
	// a limit no layer of the cache is old enough to pass, ten years

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "prune", "--older-than", "87600h").CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), "removed 0 layers")
}
