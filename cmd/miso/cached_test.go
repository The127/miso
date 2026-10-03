package main_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/The127/miso/cmd/miso"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func keyed(keys ...string) plan.Plan {
	steps := make([]plan.Step, 0, len(keys))
	for _, key := range keys {
		steps = append(steps, plan.Step{Instruction: imagefile.Run{Line: 2, Command: "make"}, Key: key})
	}

	return plan.Plan{Stages: []plan.Stage{{Base: "debian:13", Steps: steps}}}
}

func TestACacheWithoutADiskHoldsNoKeysAndBootsNothing(t *testing.T) {
	// arrange
	cache := t.TempDir()

	// act
	cached, err := main.CachedKeys(t.Context(), nil, cache, nil, keyed("9a8b7c6d5e4f"))

	// assert
	require.NoError(t, err)
	assert.Empty(t, cached)
}

func TestAPlanWithoutKeysAsksNothingOfTheCacheDisk(t *testing.T) {
	// arrange
	cache := t.TempDir()
	builder := filepath.Join(cache, "builder")
	require.NoError(t, os.Mkdir(builder, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(builder, "layers.img"), nil, 0o600))

	// act
	cached, err := main.CachedKeys(t.Context(), nil, cache, nil, keyed(""))

	// assert
	require.NoError(t, err)
	assert.Empty(t, cached)
}
