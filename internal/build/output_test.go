package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/protocol"
)

// disksOf are the disks among the requests, in their order.
func disksOf(requests []build.Request) []protocol.Disk {
	var disks []protocol.Disk
	for _, request := range requests {
		if disk, isDisk := request.Message.(protocol.Disk); isDisk {
			disks = append(disks, disk)
		}
	}

	return disks
}

// rootfsOf is the file system the last output asked for, which comes before
// the fetch of its file.
func rootfsOf(t *testing.T, requests []build.Request) protocol.Rootfs {
	t.Helper()

	rootfs, isRootfs := requests[len(requests)-2].Message.(protocol.Rootfs)
	require.True(t, isRootfs)

	return rootfs
}

func TestADiskOutputIsMadeOfTheLayersOfItsStageWithTheLayersOfTheToolsStageMisoAdds(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nRUN apt-get install htop\nOUTPUT disk os.raw\n")
	tools, image := source.Stages[0], source.Stages[1]

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	require.Len(t, disksOf(requests), 1)
	assert.Equal(t, protocol.Disk{
		Key:    image.Steps[1].Key,
		Layers: []string{image.BaseKey, image.Steps[0].Key},
		Tools:  []string{tools.BaseKey, tools.Steps[0].Key},
	}, disksOf(requests)[0])
}

func TestAnISOOutputIsADiskThatBootsFromACD(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nRUN apt-get install htop\nOUTPUT iso os.iso\n")
	tools, image := source.Stages[0], source.Stages[1]

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	require.Len(t, disksOf(requests), 1)
	assert.Equal(t, protocol.Disk{
		Key:      image.Steps[1].Key,
		Layers:   []string{image.BaseKey, image.Steps[0].Key},
		Tools:    []string{tools.BaseKey, tools.Steps[0].Key},
		ElTorito: true,
	}, disksOf(requests)[0])
}

func TestARootfsOutputIsAnExt4FileSystemOfTheLayersOfItsStageWithTheLayersOfTheToolsStageMisoAdds(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nRUN apt-get install htop\nOUTPUT rootfs os.ext4\n")
	tools, image := source.Stages[0], source.Stages[1]

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	require.Len(t, requests, 5)
	assert.Equal(t, protocol.Rootfs{
		Key:    image.Steps[1].Key,
		Layers: []string{image.BaseKey, image.Steps[0].Key},
		Tools:  []string{tools.BaseKey, tools.Steps[0].Key},
		Format: "ext4",
		Name:   "os.ext4",
	}, requests[len(requests)-2].Message)
}

func TestARootfsOutputWithAFormatAsksForThatFormat(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT rootfs --format=erofs os.erofs\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	rootfs := rootfsOf(t, requests)
	assert.Equal(t, "erofs", rootfs.Format)
}

func TestAPortableOutputIsARootfsThatIsWrappedInADisk(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT portable web.raw\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	rootfs := rootfsOf(t, requests)
	assert.Equal(t, protocol.WrapPortable, rootfs.Wrap)
	assert.Equal(t, "ext4", rootfs.Format)
}

func TestAPortableOutputWithAFormatAsksForThatFormat(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT portable --format=erofs web.raw\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	rootfs := rootfsOf(t, requests)
	assert.Equal(t, "erofs", rootfs.Format)
}

func TestASysextOutputIsARootfsThatIsWrappedInADiskAsASysext(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT sysext tools.raw\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	rootfs := rootfsOf(t, requests)
	assert.Equal(t, protocol.WrapSysext, rootfs.Wrap)
	assert.Equal(t, "tools.raw", rootfs.Name)
}

func TestAConfextOutputIsARootfsThatIsWrappedInADiskAsAConfext(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT confext app.raw\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	rootfs := rootfsOf(t, requests)
	assert.Equal(t, protocol.WrapConfext, rootfs.Wrap)
	assert.Equal(t, "app.raw", rootfs.Name)
}

func TestAConfextOutputWithAFormatAsksForThatFormat(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT confext --format=erofs app.raw\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	rootfs := rootfsOf(t, requests)
	assert.Equal(t, "erofs", rootfs.Format)
}

func TestASysextOutputWithAFormatAsksForThatFormat(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT sysext --format=erofs tools.raw\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	rootfs := rootfsOf(t, requests)
	assert.Equal(t, "erofs", rootfs.Format)
}

func TestARootfsOfAFormatMisoCannotMakeFailsAtItsLine(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT rootfs --format=btrfs os.img\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrUnknownFormat)
	assert.ErrorContains(t, err, "line 2")
	assert.ErrorContains(t, err, "btrfs")
}

func TestADiskIsFetchedIntoTheFileItsOutputNames(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT disk os.raw\n")
	output := source.Stages[1].Steps[0]

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	last := requests[len(requests)-1]
	assert.Equal(t, protocol.Fetch{Key: output.Key}, last.Message)
	assert.Equal(t, "os.raw", last.Output)
}

func TestTheFetchOfAnISOSaysItIsCheckedAsACD(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT iso os.iso\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	assert.True(t, requests[len(requests)-1].CD)
}

func TestTheFetchOfADiskCarriesTheChecksAfterItsOutput(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT disk os.raw\nCHECK command -v htop\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	last := requests[len(requests)-1]
	assert.Equal(t, []imagefile.Check{{Line: 3, Command: "command -v htop"}}, last.Checks)
}

func TestAStepBetweenAnOutputAndItsCheckLeavesTheCheckWithTheFetch(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT disk os.raw\nRUN true\nCHECK command -v htop\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	fetch := requests[len(requests)-2]
	require.IsType(t, protocol.Fetch{}, fetch.Message)
	assert.Equal(t, []imagefile.Check{{Line: 4, Command: "command -v htop"}}, fetch.Checks)
}

func TestAnOutputOfAKindMisoCannotMakeFailsAtItsLine(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT widget os.raw\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrUnknownKind)
	assert.ErrorContains(t, err, "line 2")
	assert.ErrorContains(t, err, "widget")
}

func TestADiskWithAnOptionMisoDoesNotKnowFailsAtItsLine(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT disk os.raw --kernal=6.1\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrUnknownOption)
	assert.ErrorContains(t, err, "line 2")
	assert.ErrorContains(t, err, "kernal")
}

func TestAToolsOptionOnADiskIsAnOptionMisoDoesNotKnow(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13 AS tools\nFROM debian:13\nOUTPUT disk os.raw --tools=tools\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrUnknownOption)
	assert.ErrorContains(t, err, "line 3")
	assert.ErrorContains(t, err, "--tools")
}

func TestADiskIsMadeOfThePartitionsAboveIt(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nPARTITION esp Format=vfat\nPARTITION root Format=ext4 SizeMinBytes=3G\nOUTPUT disk os.raw\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	require.Len(t, disksOf(requests), 1)
	assert.Equal(t, []protocol.Partition{
		{Name: "esp", Settings: []protocol.Setting{{Key: "Format", Value: "vfat"}}},
		{Name: "root", Settings: []protocol.Setting{{Key: "Format", Value: "ext4"}, {Key: "SizeMinBytes", Value: "3G"}}},
	}, disksOf(requests)[0].Partitions)
}

func TestADiskGetsTheCmdlineLinesAboveItJoinedByASpace(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nCMDLINE rw\nCMDLINE console=ttyS0\nOUTPUT disk os.raw\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	require.Len(t, disksOf(requests), 1)
	assert.Equal(t, "rw console=ttyS0", disksOf(requests)[0].Cmdline)
}

func TestARootfsWithAnEmptyFormatFailsAtItsLine(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT rootfs --format= os.img\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrUnknownFormat)
	assert.ErrorContains(t, err, "line 2")
}
