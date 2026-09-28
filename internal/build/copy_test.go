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

// copyingMotdAndIssue is a plan whose one step copies two files of the
// build context into /etc/ of a Debian image.
func copyingMotdAndIssue() plan.Plan {
	return plan.Plan{Stages: []plan.Stage{{
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
}

// copiesOf are the requests that copy, in their order.
func copiesOf(requests []build.Request) []build.Request {
	var copies []build.Request
	for _, request := range requests {
		if _, isCopy := request.Message.(protocol.Copy); isCopy {
			copies = append(copies, request)
		}
	}

	return copies
}

func TestACopyOfTheBuildContextBecomesARequestWithItsFiles(t *testing.T) {
	// arrange
	planned := copyingMotdAndIssue()

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	require.Len(t, copiesOf(requests), 1)
	assert.Equal(t, protocol.Copy{
		Key:         "k1",
		Layers:      []string{"base"},
		Sources:     []string{"motd", "issue"},
		Digests:     []string{"miso-context-1:aaaa", "miso-context-1:bbbb"},
		Destination: "/etc/",
	}, copiesOf(requests)[0].Message)
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
	assert.Empty(t, copiesOf(requests))
}

func TestACopyRequestKnowsItsLineOfTheBuildFile(t *testing.T) {
	// arrange
	planned := copyingMotdAndIssue()

	// act
	requests, err := build.Requests(planned, network)

	// assert
	require.NoError(t, err)
	require.Len(t, copiesOf(requests), 1)
	assert.Equal(t, 2, copiesOf(requests)[0].Line)
	assert.Equal(t, "COPY motd issue /etc/", copiesOf(requests)[0].Written)
}
