package build_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/build"
)

const checkedRootfs = "FROM debian:13\nCMDLINE root=/dev/vda rw\nCMDLINE console=ttyS0\nOUTPUT kernel vmlinuz\nOUTPUT initrd initrd.img\nOUTPUT rootfs os.ext4\nCHECK true\n"

func TestTheFetchOfACheckedRootfsNamesTheKernelAndInitrdItIsBootedWithAndItsCmdline(t *testing.T) {
	// arrange
	source := planned(t, checkedRootfs)

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	last := requests[len(requests)-1]
	assert.Equal(t, &build.BootFiles{Kernel: "vmlinuz", Initrd: "initrd.img", Cmdline: "root=/dev/vda rw console=ttyS0"}, last.Boot)
}

func TestTheFetchesOfTheKernelAndInitrdACheckedRootfsIsBootedWithAreNeeded(t *testing.T) {
	// arrange
	source := planned(t, checkedRootfs)

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	needed := map[string]bool{}
	for _, request := range requests {
		if request.Output != "" {
			needed[request.Output] = request.Needed
		}
	}

	assert.Equal(t, map[string]bool{"vmlinuz": true, "initrd.img": true, "os.ext4": false}, needed)
}

func TestTheFetchOfACheckedDiskHasNoBootFilesAndNeedsNoOtherFetch(t *testing.T) {
	// arrange
	source := planned(t, "FROM debian:13\nOUTPUT kernel vmlinuz\nOUTPUT initrd initrd.img\nOUTPUT disk os.raw\nCHECK true\n")

	// act
	requests, err := build.Requests(source, network)

	// assert
	require.NoError(t, err)
	assert.Nil(t, requests[len(requests)-1].Boot)
	for _, request := range requests {
		assert.False(t, request.Needed)
	}
}
