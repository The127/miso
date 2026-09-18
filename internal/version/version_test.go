package version_test

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/version"
)

func TestAReleasedBinaryShowsItsVersion(t *testing.T) {
	// arrange
	info := &debug.BuildInfo{Main: debug.Module{Version: "v0.3.1"}}

	// act
	shown := version.Of(info)

	// assert
	assert.Equal(t, "v0.3.1", shown)
}
