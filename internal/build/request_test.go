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

// network is the host's network in these tests.
var network = protocol.Network{Card: "52:54:00:6d:69:73", IPv4: protocol.Family{Address: "10.0.2.15/24", Gateway: "10.0.2.2"}}

// runsOf are the runs among the requests, in their order.
func runsOf(requests []protocol.Message) []protocol.Run {
	var runs []protocol.Run
	for _, request := range requests {
		if run, isRun := request.(protocol.Run); isRun {
			runs = append(runs, run)
		}
	}

	return runs
}

func TestARunBecomesARequestWithItsCommand(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps:      []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Equal(t, "echo hi", runsOf(requests)[0].Command)
}

func TestAnOnlineRunCarriesTheNetworkOfTheHost(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps:      []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "apt-get update"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Equal(t, &network, runsOf(requests)[0].Network)
}

func TestAnOfflineRunBecomesAnOfflineRequest(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps:      []plan.Step{{Instruction: imagefile.Run{Line: 2, Offline: true, Command: "make test"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Nil(t, runsOf(requests)[0].Network)
}

func TestARequestCarriesTheKeyOfItsStep(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps:      []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	require.Len(t, runsOf(requests), 1)
	assert.Equal(t, "k1", runsOf(requests)[0].Key)
}

func TestAPlanWithABaseToFetchHasNoRequests(t *testing.T) {
	// arrange
	planned := plan.Plan{Downloads: []string{"debian-13"}, Stages: []plan.Stage{{
		Base:  "debian-13",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, BuiltOn: []string{""}}},
	}}}

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, build.ErrNotFetched)
	assert.ErrorContains(t, err, "debian-13")
	assert.Empty(t, requests)
}
