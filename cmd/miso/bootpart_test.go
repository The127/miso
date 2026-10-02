//go:build kvm

package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// the setup header of a bzImage says HdrS at this place
const (
	bzImageMagic       = "HdrS"
	bzImageMagicOffset = 0x202
)

func TestKernelAndInitrdOutputsAreTheFilesOfTheImageOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	out := t.TempDir()
	imagefile := "FROM debian:sid\nOUTPUT kernel vmlinuz\nOUTPUT initrd initrd.img\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(imagefile), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", "-o", out, dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	kernel, err := os.ReadFile(filepath.Join(out, "vmlinuz"))
	require.NoError(t, err)
	require.Greater(t, len(kernel), bzImageMagicOffset+len(bzImageMagic))
	assert.Equal(t, bzImageMagic, string(kernel[bzImageMagicOffset:bzImageMagicOffset+len(bzImageMagic)]))
	initrd, err := os.Stat(filepath.Join(out, "initrd.img"))
	require.NoError(t, err)
	assert.Positive(t, initrd.Size())
}

// the ELF magic and the machine of an x86-64 executable
const (
	elfMagicBytes = "\x7fELF"
	elfMachineAt  = 18
	elfX86_64     = 62
)

func TestAnUnpackedKernelOutputIsTheELFFileOfTheImagesKernelOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	out := t.TempDir()
	imagefile := "FROM debian:sid\nOUTPUT kernel --elf vmlinux\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(imagefile), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", "-o", out, dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	kernel, err := os.ReadFile(filepath.Join(out, "vmlinux"))
	require.NoError(t, err)
	require.Greater(t, len(kernel), elfMachineAt+2)
	assert.Equal(t, elfMagicBytes, string(kernel[:len(elfMagicBytes)]))
	assert.Equal(t, byte(elfX86_64), kernel[elfMachineAt])
}
