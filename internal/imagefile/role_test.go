package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/imagefile"
)

func TestRunEnvAndCopyBuildTheRootFileSystem(t *testing.T) {
	// arrange
	instructions := []imagefile.Instruction{imagefile.Run{}, imagefile.Env{}, imagefile.Copy{}}

	// act
	roles := []imagefile.Role{
		imagefile.RoleOf(instructions[0]),
		imagefile.RoleOf(instructions[1]),
		imagefile.RoleOf(instructions[2]),
	}

	// assert
	assert.Equal(t, []imagefile.Role{imagefile.Builds, imagefile.Builds, imagefile.Builds}, roles)
}

func TestPartitionAndCmdlineSaySomethingAboutTheDisks(t *testing.T) {
	// arrange
	partition := imagefile.Partition{}
	cmdline := imagefile.Cmdline{}

	// act
	roles := []imagefile.Role{imagefile.RoleOf(partition), imagefile.RoleOf(cmdline)}

	// assert
	assert.Equal(t, []imagefile.Role{imagefile.ForDisks, imagefile.ForDisks}, roles)
}

func TestOutputAndCheckMakeArtifacts(t *testing.T) {
	// arrange
	output := imagefile.Output{}
	check := imagefile.Check{}

	// act
	roles := []imagefile.Role{imagefile.RoleOf(output), imagefile.RoleOf(check)}

	// assert
	assert.Equal(t, []imagefile.Role{imagefile.Artifact, imagefile.Artifact}, roles)
}
