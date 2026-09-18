package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

func TestAStageOnAnImageImportsItsBaseFirst(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps:      []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.NotEmpty(t, requests)
	assert.Equal(t, protocol.Import{Key: "base", Digest: "sha256:image"}, requests[0])
}

func TestAStageOnScratchImportsNothing(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:    "scratch",
		BaseKey: "s",
		Steps:   []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"s"}}},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, requests, 1)
	assert.IsType(t, protocol.Run{}, requests[0])
}

func TestAStageOnAnEmptyScratchStageImportsNothing(t *testing.T) {
	// arrange
	source := planned(t, "FROM scratch AS empty\nFROM empty\nRUN x\n")

	// act
	requests, err := build.Requests(source)

	// assert
	require.NoError(t, err)
	require.Len(t, requests, 1)
	assert.IsType(t, protocol.Run{}, requests[0])
}
