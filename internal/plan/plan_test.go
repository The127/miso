package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

const anyAgent = "agent"

func keys(t *testing.T, stages []imagefile.Stage, agent string, context plan.Context, bases plan.Bases) [][]string {
	t.Helper()

	planned, err := plan.New(stages, agent, context, bases)
	require.NoError(t, err)

	var found [][]string
	for _, stage := range planned.Stages {
		var stageKeys []string
		for _, step := range stage.Steps {
			stageKeys = append(stageKeys, step.Key)
		}

		found = append(found, stageKeys)
	}

	return found
}

func lastKey(t *testing.T, found [][]string) string {
	t.Helper()

	require.NotEmpty(t, found)
	stage := found[len(found)-1]
	require.NotEmpty(t, stage)

	return stage[len(stage)-1]
}

func TestAnInvalidBuildFileGetsNoPlan(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nCOPY --from=nope a /b\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, noImages)

	// assert
	assert.ErrorIs(t, err, plan.ErrUnknownStage)
	assert.Empty(t, planned.Stages)
}

func TestAPlanKeepsTheBaseDigestItWasKeyedWith(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nRUN true\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, images{"debian:sid": "sha256:old"})

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 1)
	assert.Equal(t, "sha256:old", planned.Stages[0].BaseDigest)
}

func TestAPlannedStepKnowsItsInstruction(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nRUN true\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, noImages)

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 1)
	require.Len(t, planned.Stages[0].Steps, 1)
	assert.Equal(t, stages[0].Instructions[0], planned.Stages[0].Steps[0].Instruction)
	assert.NotEmpty(t, planned.Stages[0].Steps[0].Key)
}

func TestAPlannedCopyKeepsTheDigestsOfItsFiles(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nCOPY motd issue /etc/\n")

	// act
	planned, err := plan.New(stages, anyAgent, files{"motd": "hello", "issue": "welcome"}, noImages)

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 1)
	require.Len(t, planned.Stages[0].Steps, 1)
	assert.Equal(t, []plan.File{{Path: "motd", Digest: "hello"}, {Path: "issue", Digest: "welcome"}}, planned.Stages[0].Steps[0].Files)
}

func TestAPlannedRunAfterAnOutputWasBuiltOnTheRunBeforeIt(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nRUN make\nOUTPUT disk os.img\nRUN make install\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, noImages)

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 1)
	steps := planned.Stages[0].Steps
	require.Len(t, steps, 3)
	assert.Equal(t, []string{steps[0].Key}, steps[2].BuiltOn)
}

func TestAPlannedCheckWasBuiltOnTheOutputsBeforeIt(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nOUTPUT disk os.img\nOUTPUT portable app.raw\nCHECK true\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, noImages)

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 1)
	steps := planned.Stages[0].Steps
	require.Len(t, steps, 3)
	assert.Equal(t, []string{steps[0].Key, steps[1].Key}, steps[2].BuiltOn)
}

func TestAPlannedStageKeepsItsNameAndItsBase(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid AS build\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, debianImages)

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 1)
	assert.Equal(t, "build", planned.Stages[0].Name)
	assert.Equal(t, "debian:sid", planned.Stages[0].Base)
}
