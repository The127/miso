package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

func TestARunOnScratchBecomesARequestWithItsCommand(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "echo hi"}, Key: "k1", BuiltOn: []string{"base"}}},
	}}}

	// act
	requests := build.Requests(planned)

	// assert
	assert.Equal(t, []protocol.Run{{Command: "echo hi"}}, requests)
}
