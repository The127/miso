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

func TestACopyOfTheBuildContextBecomesARequestWithItsFiles(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{{
			Instruction: imagefile.Copy{Line: 2, Sources: []string{"motd", "issue"}, Destination: "/etc/"},
			Key:         "k1",
			BuiltOn:     []string{"base"},
			Files:       []plan.File{{Path: "motd", Digest: "miso-context-1:aaaa"}, {Path: "issue", Digest: "miso-context-1:bbbb"}},
		}},
	}}}

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	require.Len(t, requests, 2)
	assert.Equal(t, protocol.Copy{
		Key:         "k1",
		Layers:      []string{"base"},
		Sources:     []string{"motd", "issue"},
		Digests:     []string{"miso-context-1:aaaa", "miso-context-1:bbbb"},
		Destination: "/etc/",
	}, requests[1].Message)
}

func TestACopyFromAStageTakesNothingFromTheBuildContext(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:       "debian:13",
		BaseDigest: "sha256:image",
		BaseKey:    "base",
		Steps: []plan.Step{{
			Instruction: imagefile.Copy{Line: 2, From: "build", Sources: []string{"rootfs"}, Destination: "/"},
			Key:         "k1",
			BuiltOn:     []string{"base"},
		}},
	}}}

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	for _, request := range requests {
		assert.IsNotType(t, protocol.Copy{}, request.Message)
	}
}
