package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/plan"
)

func TestACmdlineLeavesTheLayersOfTheStepsAfterItAlone(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nCMDLINE rw\nRUN true\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, noImages)

	// assert
	require.NoError(t, err)
	stage := planned.Stages[0]
	require.Len(t, stage.Steps, 2)
	assert.Equal(t, []string{stage.BaseKey}, stage.Steps[1].BuiltOn)
}

func TestADiskGetsADifferentKeyWhenItsCmdlineChanges(t *testing.T) {
	// arrange
	serial := parse(t, "FROM scratch\nCMDLINE console=ttyS0\nOUTPUT disk os.img\n")
	quiet := parse(t, "FROM scratch\nCMDLINE quiet\nOUTPUT disk os.img\n")

	// act
	serialKeys := keys(t, serial, anyAgent, noFiles, toolsImages)
	quietKeys := keys(t, quiet, anyAgent, noFiles, toolsImages)

	// assert
	assert.NotEqual(t, lastKey(t, serialKeys), lastKey(t, quietKeys))
}
