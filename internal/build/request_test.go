package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func TestARunOnScratchBecomesARequestWithItsCommand(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, requests, 1)
	assert.Equal(t, "echo hi", requests[0].Command)
}

func TestARequestCarriesTheKeyOfItsStep(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.NoError(t, err)
	require.Len(t, requests, 1)
	assert.Equal(t, "k1", requests[0].Key)
}

func TestAPlanWithABaseToFetchHasNoRequests(t *testing.T) {
	// arrange
	planned := plan.Plan{Downloads: []string{"debian-13"}, Stages: []plan.Stage{{
		Base:  "debian-13",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, BuiltOn: []string{""}}},
	}}}

	// act
	requests, err := build.Requests(planned)

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, build.ErrNotFetched)
	assert.ErrorContains(t, err, "debian-13")
	assert.Empty(t, requests)
}
