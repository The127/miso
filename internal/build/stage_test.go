package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

// fetched says every image is fetched, with one digest for all.
type fetched struct{}

func (fetched) Digest(string) (string, error) { return "sha256:image", nil }

// noContext is a build context with no files in it.
type noContext struct{}

func (noContext) Digest(string) (string, error) { return "", nil }

func planned(t *testing.T, source string) plan.Plan {
	t.Helper()

	stages, err := imagefile.Parse(source)
	require.NoError(t, err)
	planned, err := plan.New(stages, "agent", noContext{}, fetched{})
	require.NoError(t, err)

	return planned
}

func TestAStageOnAStageIsBuiltOnTheLayersOfThatStage(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13 AS one\nRUN a\nFROM one\nRUN b\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 2)
	assert.Equal(t, []string{source.Stages[0].BaseKey, runsOf(requests)[0].Key}, runsOf(requests)[1].Layers)
}
