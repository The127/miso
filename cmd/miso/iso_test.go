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

// bootableFromCD is a build file of an ISO. Its image brings a kernel with
// an initrd that finds a root on a CD, which Debian's dracut does only with
// systemd's loop unit and systemd-dissect added, and a read-only root.
const bootableFromCD = `FROM debian:sid
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y full-upgrade && \
    DEBIAN_FRONTEND=noninteractive apt-get purge -y 'linux-image-*cloud*' && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends linux-image-amd64 dracut systemd-boot-efi systemd-container
RUN v=$(ls /usr/lib/modules | sort -V | tail -1) && \
    dracut --force --no-hostonly --add-drivers "erofs sr_mod loop" \
      --install "/usr/bin/systemd-dissect /usr/lib/systemd/system/systemd-loop@.service /usr/lib/udev/rules.d/99-systemd.rules" \
      --kver "$v" "/boot/initrd.img-$v"
RUN : > /etc/fstab && \
    mkdir -p /etc/kernel && \
    printf 'ro rootfstype=erofs systemd.volatile=overlay console=ttyS0\n' > /etc/kernel/cmdline
PARTITION esp Type=esp Format=vfat CopyFiles=/efi:/ SizeMinBytes=256M SizeMaxBytes=256M
PARTITION root Type=root Format=erofs CopyFiles=/:/ ExcludeFiles=/efi/ Minimize=best ReadOnly=yes
OUTPUT iso os.iso
CHECK test -b /dev/sr0
`

func TestAnISOIsCheckedBootedFromACDOnTheBuilderKernel(t *testing.T) {
	// arrange
	dir := t.TempDir()
	out := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(bootableFromCD), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "build", "-o", out, dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.NoError(t, err, string(said))
	assert.FileExists(t, filepath.Join(out, "os.iso"))
}
