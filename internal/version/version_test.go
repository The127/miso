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

func TestADevelopmentBuildShowsItsCommit(t *testing.T) {
	// arrange
	info := &debug.BuildInfo{
		Main:     debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "a436a5d37ff86562c5bc73ae75e1092b147b1968"}},
	}

	// act
	shown := version.Of(info)

	// assert
	assert.Equal(t, "a436a5d37ff8", shown)
}
