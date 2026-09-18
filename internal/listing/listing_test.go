package listing_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/listing"
	"github.com/The127/miso/internal/plan"
)

func TestAStepShowsItsLineItsKeyAndItsCommand(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "apt-get install -y vim"}, Key: "9a8b7c6d5e4f"}},
	}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM scratch\n   2  9a8b7c6d5e4f  RUN apt-get install -y vim\n", out.String())
}
