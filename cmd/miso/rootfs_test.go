//go:build kvm

package main_test

import (
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// the magic number of the ext file systems, at its place in the superblock:
// the superblock starts at 1024 and the magic is 56 bytes into it
const (
	extMagic       = 0xEF53
	extMagicOffset = 1024 + 56
)

func TestARootfsOutputIsAnExt4FileSystemOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	out := t.TempDir()
	imagefile := "FROM debian:sid\nRUN apt-get update && apt-get install -y --no-install-recommends htop\nOUTPUT rootfs os.ext4\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(imagefile), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", "-o", out, dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	file, err := os.Open(filepath.Join(out, "os.ext4"))
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	magic := make([]byte, 2)
	_, err = file.ReadAt(magic, extMagicOffset)
	require.NoError(t, err)
	assert.Equal(t, uint16(extMagic), binary.LittleEndian.Uint16(magic))
}

// the magic number of erofs, at its place in the file: the superblock
// starts 1024 bytes in and the magic is its first field
const (
	erofsMagic       = 0xE0F5E1E2
	erofsMagicOffset = 1024
)

func TestAnErofsRootfsOutputIsAnErofsFileSystemOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	out := t.TempDir()
	imagefile := "FROM debian:sid\nRUN apt-get update && apt-get install -y --no-install-recommends htop\nOUTPUT rootfs --format=erofs os.erofs\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(imagefile), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", "-o", out, dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	file, err := os.Open(filepath.Join(out, "os.erofs"))
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	magic := make([]byte, 4)
	_, err = file.ReadAt(magic, erofsMagicOffset)
	require.NoError(t, err)
	assert.Equal(t, uint32(erofsMagic), binary.LittleEndian.Uint32(magic))
}

// bootableRootfs is a build file of a rootfs that a kernel boots directly.
// Its initrd must not be bound to the disk of the cloud image, so it is made
// again for any machine, with the driver of the virtual disk.
const bootableRootfs = `FROM debian:sid
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y full-upgrade && \
    DEBIAN_FRONTEND=noninteractive apt-get purge -y 'linux-image-*cloud*' && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends linux-image-amd64 dracut
RUN v=$(ls /usr/lib/modules | sort -V | tail -1) && \
    dracut --force --no-hostonly --add-drivers "virtio_blk ext4" --kver "$v" "/boot/initrd.img-$v"
RUN : > /etc/fstab
CMDLINE root=/dev/vda rootfstype=ext4 rw console=ttyS0
OUTPUT kernel vmlinuz
OUTPUT initrd initrd.img
OUTPUT rootfs os.ext4
CHECK test -b /dev/vda
`

func TestARootfsIsCheckedBootedWithItsKernelAndInitrdOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	out := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(bootableRootfs), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", "-o", out, dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.FileExists(t, filepath.Join(out, "os.ext4"))
	assert.FileExists(t, filepath.Join(out, "vmlinuz"))
	assert.FileExists(t, filepath.Join(out, "initrd.img"))
}

func TestARootfsWhoseCheckFailsIsNotDeliveredOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	out := t.TempDir()
	failing := strings.Replace(bootableRootfs, "CHECK test -b /dev/vda", "CHECK test -b /dev/not-there", 1)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(failing), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", "-o", out, dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.Error(t, err, string(said))
	assert.Contains(t, string(said), "CHECK")
	assert.NoFileExists(t, filepath.Join(out, "os.ext4"))
}

func TestARootfsIsCheckedWithoutAnOutputDirectoryOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(bootableRootfs), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	assert.NoError(t, err, string(said))
}
