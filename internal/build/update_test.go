package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/protocol"
)

const updateSource = "FROM debian:13\nPARTITION esp Type=esp Format=vfat\nPARTITION root Type=root Format=ext4 SplitName=root\nOUTPUT update --version=1.2 updates\n"

func TestAnUpdateOutputAsksForASplitDisk(t *testing.T) {
	// arrange
	source := planned(t, updateSource)

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	disks := disksOf(requests)
	require.Len(t, disks, 1)
	assert.True(t, disks[0].Split)
}

func TestAnUpdateOutputFetchesTheFileOfEachPartitionWithASplitNameIntoItsDirectory(t *testing.T) {
	// arrange
	source := planned(t, updateSource)
	image := source.Stages[1]

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	fetch := requests[len(requests)-2]
	assert.Equal(t, protocol.Fetch{Key: image.Steps[len(image.Steps)-1].Key, File: "disk.root.raw"}, fetch.Message)
	assert.Equal(t, "updates/root_1.2.raw", fetch.Output)
}

func TestAnUpdateOutputFetchesTheUKIAfterThePartitionsWithTheVersionInItsName(t *testing.T) {
	// arrange
	source := planned(t, updateSource)
	image := source.Stages[1]

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	fetch := requests[len(requests)-1]
	assert.Equal(t, protocol.Fetch{Key: image.Steps[len(image.Steps)-1].Key, File: "uki.efi"}, fetch.Message)
	assert.Equal(t, "updates/uki_1.2.efi", fetch.Output)
}

func TestTheFilesOfAnUpdateAreListedForSysupdate(t *testing.T) {
	// arrange
	source := planned(t, updateSource)

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	fetches := requests[len(requests)-2:]
	assert.True(t, fetches[0].Listed)
	assert.True(t, fetches[1].Listed)
}

func TestAnUpdateOutputWithASplitNameOfOnlyADashFailsNamingIt(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nPARTITION root Type=root SplitName=-\nOUTPUT update --version=1.2 updates\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrBadSplitName)
	assert.ErrorContains(t, err, "SplitName=-")
}

func TestAnUpdateOutputWhoseTwoPartitionsHaveOneSplitNameFailsNamingIt(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nPARTITION a Type=root SplitName=root\nPARTITION b Type=root SplitName=root\nOUTPUT update --version=1.2 updates\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrDuplicateSplitName)
	assert.ErrorContains(t, err, "root")
}

func TestAnUpdateOutputWithoutAVersionFailsAtItsLine(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nPARTITION root Type=root SplitName=root\nOUTPUT update updates\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrNoVersion)
	assert.ErrorContains(t, err, "line 3")
}

func TestAnUpdateOutputWithAVersionSysupdateCannotReadFailsNamingIt(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nPARTITION root Type=root SplitName=root\nOUTPUT update --version=1/2 updates\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrBadVersion)
	assert.ErrorContains(t, err, "1/2")
}

func TestAnUpdateOutputWithASplitNameThatIsNoPlainNameFailsNamingIt(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nPARTITION root Type=root SplitName=../root\nOUTPUT update --version=1.2 updates\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrBadSplitName)
	assert.ErrorContains(t, err, "../root")
}

func TestAnUpdateOutputWithoutAPartitionThatIsSplitFailsAtItsLine(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nPARTITION root Type=root\nOUTPUT update --version=1.2 updates\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrNothingToShip)
	assert.ErrorContains(t, err, "line 3")
}
