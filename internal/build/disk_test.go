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

func TestADiskOutputIsMadeOfTheLayersOfItsStageWithTheLayersOfItsTools(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13 AS tools\nRUN apt-get install systemd-repart\nFROM debian:13\nRUN apt-get install htop\nOUTPUT disk os.raw --tools=tools\n")
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

func TestADiskIsFetchedIntoTheFileItsOutputNames(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13 AS tools\nFROM debian:13\nOUTPUT disk os.raw --tools=tools\n")
	output := source.Stages[1].Steps[0]

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	last := requests[len(requests)-1]
	assert.Equal(t, protocol.Fetch{Key: output.Key}, last.Message)
	assert.Equal(t, "os.raw", last.Output)
}

func TestTheFetchOfADiskCarriesTheChecksAfterItsOutput(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13 AS tools\nFROM debian:13\nOUTPUT disk os.raw --tools=tools\nCHECK command -v htop\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	last := requests[len(requests)-1]
	assert.Equal(t, []imagefile.Check{{Line: 4, Command: "command -v htop"}}, last.Checks)
}

func TestAnOutputOfAKindMisoCannotMakeFailsAtItsLine(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT portable os.raw\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrUnknownKind)
	assert.ErrorContains(t, err, "line 2")
	assert.ErrorContains(t, err, "portable")
}

func TestADiskWithAnOptionMisoDoesNotKnowFailsAtItsLine(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13 AS tools\nFROM debian:13\nOUTPUT disk os.raw --tools=tools --kernal=6.1\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrUnknownOption)
	assert.ErrorContains(t, err, "line 3")
	assert.ErrorContains(t, err, "kernal")
}

func TestADiskWithoutToolsFailsAtItsLine(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT disk os.raw\n")

	// act
	_, err := build.Requests(source, network)

	// assert
	require.ErrorIs(t, err, build.ErrNoTools)
	assert.ErrorContains(t, err, "line 2")
}
