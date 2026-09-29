//go:build kvm

package main_test

import (
	"crypto/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestABuildRunsItsStepOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	// a word of its own, or a cache from an earlier run holds the step and
	// the build prints nothing
	word := rand.Text()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN echo "+word+"\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), word+"\n")
}

func TestAFailingStepOnTheBuilderKernelNamesItsLineOfTheBuildFile(t *testing.T) {
	// arrange
	dir := t.TempDir()
	file := filepath.Join(dir, "Imagefile")
	require.NoError(t, os.WriteFile(file, []byte("FROM debian:13\nRUN false\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.Error(t, err)
	assert.Contains(t, string(said), file+":2: RUN false: command failed")
}

func TestACopiedFileOfTheContextIsInTheImageOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	// content of its own, or a cache from an earlier run holds both steps
	// and the build prints nothing
	motd := "hello " + rand.Text() + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "motd"), []byte(motd), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nCOPY motd /etc/motd\nRUN cat /etc/motd\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), motd)
}

func TestAFileCopiedFromAnEarlierStageIsInTheImageOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	// a word of its own, or a cache from an earlier run holds every step and
	// the build prints nothing
	word := rand.Text()
	build := "FROM debian:13 AS a\nRUN echo " + word + " > /x\nFROM debian:13\nCOPY --from=a /x /y\nRUN cat /y\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(build), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), word+"\n")
}

func TestABuildOnAHostThatCannotKeepVsockPrivateSaysSoAndStillRunsOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	// a word of its own, or a cache from an earlier run holds the step and
	// the build prints nothing
	word := rand.Text()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN echo "+word+"\n"), 0o600))
	// a host that refuses miso a user namespace, as some distributions do
	refusing := `echo 0 > /proc/sys/user/max_user_namespaces && exec "$0" build "$1"`

	// act
	said, err := exec.CommandContext(t.Context(), "unshare", "--user", "--map-root-user", "sh", "-c", refusing, miso(t), dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), "vsock cannot be kept private on this host")
	assert.Contains(t, string(said), word+"\n")
}

// bootable is a build file of a disk: its image brings a kernel, its
// partitions are lines of its own, and miso's tools make the disk of them.
const bootable = `FROM debian:sid
RUN apt-get update && apt-get install -y --no-install-recommends systemd-boot-efi htop
RUN : > /etc/fstab && \
    mkdir -p /etc/kernel && \
    printf 'rw console=ttyS0\n' > /etc/kernel/cmdline

PARTITION esp Type=esp Format=vfat CopyFiles=/efi:/ SizeMinBytes=256M SizeMaxBytes=256M
PARTITION root Type=root Format=ext4 CopyFiles=/:/ SizeMinBytes=3G
OUTPUT disk os.raw
`

func TestADiskOutputIsWrittenSparseIntoTheOutputDirectoryOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	out := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(bootable), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", "-o", out, dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	disk, err := os.ReadFile(filepath.Join(out, "os.raw"))
	require.NoError(t, err)
	require.Greater(t, len(disk), 520)
	assert.Equal(t, "EFI PART", string(disk[512:520]))
	info, err := os.Stat(filepath.Join(out, "os.raw"))
	require.NoError(t, err)
	stat, isStat := info.Sys().(*syscall.Stat_t)
	require.True(t, isStat)
	assert.Less(t, stat.Blocks*512, info.Size())
}

func TestADiskWhoseChecksPassIsWrittenOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	out := t.TempDir()
	checked := strings.Replace(bootable, "OUTPUT disk os.raw\n", "OUTPUT disk os.raw\nCHECK command -v htop\n", 1)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(checked), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", "-o", out, dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.FileExists(t, filepath.Join(out, "os.raw"))
}

func TestADiskWhoseCheckFailsFailsTheBuildAtItsLineAndIsNotWrittenOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	out := t.TempDir()
	file := filepath.Join(dir, "Imagefile")
	failing := strings.Replace(bootable, "OUTPUT disk os.raw\n", "OUTPUT disk os.raw\nCHECK false\n", 1)
	require.NoError(t, os.WriteFile(file, []byte(failing), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", "-o", out, dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.Error(t, err)
	assert.Contains(t, string(said), file+":10: CHECK false: exit code 1")
	assert.NoFileExists(t, filepath.Join(out, "os.raw"))
}

func TestAFailingCheckFailsABuildThatWritesNoOutputsOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	file := filepath.Join(dir, "Imagefile")
	failing := strings.Replace(bootable, "OUTPUT disk os.raw\n", "OUTPUT disk os.raw\nCHECK false\n", 1)
	require.NoError(t, os.WriteFile(file, []byte(failing), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.Error(t, err)
	assert.Contains(t, string(said), file+":10: CHECK false: exit code 1")
}
