package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

func TestADiskWithoutToolsGetsAToolsStageOnDebianBeforeItsOwn(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nOUTPUT disk os.img\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, toolsImages)

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 2)
	assert.Equal(t, "debian:sid", planned.Stages[0].Base)
	assert.Equal(t, "scratch", planned.Stages[1].Base)
}

func TestADiskInALaterStageGetsTheToolsStageBeforeAllStages(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch AS base\nRUN true\nFROM base\nOUTPUT disk os.img\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, toolsImages)

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 3)
	assert.Equal(t, "debian:sid", planned.Stages[0].Base)
	assert.Equal(t, "scratch", planned.Stages[1].Base)
}

func TestAnIsoWithoutToolsGetsTheToolsStage(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nOUTPUT iso os.iso\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, toolsImages)

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 2)
	assert.Equal(t, "debian:sid", planned.Stages[0].Base)
}

func TestTheToolsStageInstallsTheTools(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nOUTPUT disk os.img\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, toolsImages)

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages[0].Steps, 1)
	install, isRun := planned.Stages[0].Steps[0].Instruction.(imagefile.Run)
	require.True(t, isRun)
	assert.Contains(t, install.Command, "systemd-repart")
	assert.Contains(t, install.Command, "systemd-ukify")
}

func TestADiskWithoutToolsGetsADifferentKeyWhenTheToolsBaseChanges(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nOUTPUT disk os.img\n")

	// act
	oldKeys := keys(t, stages, anyAgent, noFiles, images{"debian:sid": "sha256:old"})
	newKeys := keys(t, stages, anyAgent, noFiles, images{"debian:sid": "sha256:new"})

	// assert
	assert.NotEqual(t, lastKey(t, oldKeys), lastKey(t, newKeys))
}

func TestABuildFileWithoutADiskOrIsoGetsNoToolsStage(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nOUTPUT portable app.raw\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, noImages)

	// assert
	require.NoError(t, err)
	assert.Len(t, planned.Stages, 1)
	assert.Empty(t, planned.Downloads)
}

func TestTheToolsStageInstallsWhatAnIsoNeeds(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nOUTPUT iso os.iso\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, toolsImages)

	// assert
	require.NoError(t, err)
	install, isRun := planned.Stages[0].Steps[0].Instruction.(imagefile.Run)
	require.True(t, isRun)
	assert.Contains(t, install.Command, "erofs-utils")
}

func TestTheToolsComeFromAFixedDayOfTheDebianArchive(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nOUTPUT disk os.img\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, toolsImages)

	// assert
	require.NoError(t, err)
	install, isRun := planned.Stages[0].Steps[0].Instruction.(imagefile.Run)
	require.True(t, isRun)
	assert.Regexp(t, `https://snapshot\.debian\.org/archive/debian/\d{8}T\d{6}Z/`, install.Command)
}

func TestTheToolsStageIsOnTheLineOfTheOutputThatNeedsIt(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nRUN true\nOUTPUT disk os.img\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, toolsImages)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 3, planned.Stages[0].Line)
	install, isRun := planned.Stages[0].Steps[0].Instruction.(imagefile.Run)
	require.True(t, isRun)
	assert.Equal(t, 3, install.Line)
}

var toolsImages = images{"debian:sid": "sha256:sid"}
