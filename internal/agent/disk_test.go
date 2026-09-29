//go:build vmtest

package agent_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/kernel"
	"github.com/The127/miso/internal/protocol"
)

func TestADiskOfAnImageWithoutAKernelFailsNamingIt(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	strip := protocol.Run{Key: "stripped", Layers: []string{"base"}, Command: "rm -r /boot/vmlinuz-* /usr/lib/modules"}
	code, err := worker.Run(context.Background(), strip, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	disk := protocol.Disk{Key: "disk", Layers: []string{"base", "stripped"}, Tools: []string{"base"}}

	// act
	err = worker.Disk(context.Background(), disk, io.Discard)

	// assert
	assert.ErrorIs(t, err, kernel.ErrNoKernel)
}

// bootable adds to the base the systemd-boot an image brings, and answers
// the layers of that image.
func bootable(t *testing.T, worker *agent.Agent) []string {
	t.Helper()

	boot := "mkdir -p /usr/lib/systemd/boot/efi && echo loader > /usr/lib/systemd/boot/efi/systemd-bootx64.efi"
	run := protocol.Run{Key: "boot", Layers: []string{"base"}, Command: boot}
	code, err := worker.Run(context.Background(), run, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)

	return []string{"base", "boot"}
}

// writesDisk is a systemd-repart that writes disk into the last path it is
// given.
const writesDisk = `for last; do :; done
echo disk > "$last"`

// ukify is a ukify that writes its arguments into its output.
const ukify = `for arg; do case $arg in --output=*) output=${arg#--output=} ;; esac; done
echo "$@" > "$output"`

// fakeTools adds to the base a systemd-repart that runs the script and a
// ukify, and answers the layers of those tools.
func fakeTools(t *testing.T, worker *agent.Agent, repart string) []string {
	t.Helper()

	tool := "cat > /usr/local/bin/%s <<'EOF'\n#!/bin/sh\n%s\nEOF\nchmod 755 /usr/local/bin/%[1]s\n"
	command := fmt.Sprintf(tool, "systemd-repart", repart) + fmt.Sprintf(tool, "ukify", ukify)
	run := protocol.Run{Key: "tools", Layers: []string{"base"}, Command: command}
	code, err := worker.Run(context.Background(), run, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)

	return []string{"base", "tools"}
}

func TestADiskIsWhatItsToolsWriteKeptAsTheLayerOfItsKey(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	disk := protocol.Disk{Key: "disk", Layers: bootable(t, worker), Tools: fakeTools(t, worker, writesDisk)}

	// act
	err := worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "disk", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "disk\n", string(written))
}

func TestADiskIsMadeFromTheImageAsRepartsRoot(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	mark := rand.Text()
	image := protocol.Run{Key: "image", Layers: bootable(t, worker), Command: "echo " + mark + " > /etc/miso-image"}
	code, err := worker.Run(context.Background(), image, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	repart := `for arg; do case $arg in --root=*) root=${arg#--root=} ;; esac; last=$arg; done
cat "$root/etc/miso-image" > "$last"`
	disk := protocol.Disk{Key: "disk", Layers: []string{"base", "boot", "image"}, Tools: fakeTools(t, worker, repart)}

	// act
	err = worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "disk", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, mark+"\n", string(written))
}

func TestTheImagesSystemdBootIsTheFallbackLoaderOfTheESP(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	mark := rand.Text()
	boot := "mkdir -p /usr/lib/systemd/boot/efi && echo " + mark + " > /usr/lib/systemd/boot/efi/systemd-bootx64.efi"
	image := protocol.Run{Key: "image", Layers: []string{"base"}, Command: boot}
	code, err := worker.Run(context.Background(), image, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	repart := `for arg; do case $arg in --root=*) root=${arg#--root=} ;; esac; last=$arg; done
cat "$root/efi/EFI/BOOT/BOOTX64.EFI" > "$last"`
	disk := protocol.Disk{Key: "disk", Layers: []string{"base", "image"}, Tools: fakeTools(t, worker, repart)}

	// act
	err = worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "disk", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, mark+"\n", string(written))
}

func TestTheESPHoldsAUKIOfTheImagesKernelNamedByItsVersion(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	kernel := "mkdir -p /usr/lib/modules/99.0 && touch /usr/lib/modules/99.0/vmlinuz /usr/lib/modules/99.0/initrd"
	image := protocol.Run{Key: "image", Layers: bootable(t, worker), Command: kernel}
	code, err := worker.Run(context.Background(), image, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	repart := `for arg; do case $arg in --root=*) root=${arg#--root=} ;; esac; last=$arg; done
cat "$root/efi/EFI/Linux/99.0.efi" > "$last"`
	disk := protocol.Disk{Key: "disk", Layers: []string{"base", "boot", "image"}, Tools: fakeTools(t, worker, repart)}

	// act
	err = worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "disk", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "build"+
		" --linux=/run/miso/image/usr/lib/modules/99.0/vmlinuz"+
		" --initrd=/run/miso/image/usr/lib/modules/99.0/initrd"+
		" --stub=/run/miso/image/usr/lib/systemd/boot/efi/linuxx64.efi.stub"+
		" --os-release=@/run/miso/image/etc/os-release"+
		" --uname=99.0"+
		" --output=/run/miso/esp/EFI/Linux/99.0.efi\n", string(written))
}

func TestRepartWritesANewDiskOfflineWith512ByteSectors(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	repart := `for last; do :; done
echo "$@" > "$last"`
	disk := protocol.Disk{Key: "disk", Layers: bootable(t, worker), Tools: fakeTools(t, worker, repart)}

	// act
	err := worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "disk", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "--dry-run=no --root=/run/miso/image --offline=yes --sector-size=512 --empty=create --size=auto /run/miso/out/disk.raw\n", string(written))
}

func TestTheToolsCannotOpenADeviceOfTheImage(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	image := protocol.Run{Key: "image", Layers: bootable(t, worker), Command: "mknod /etc/miso-null c 1 3"}
	code, err := worker.Run(context.Background(), image, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	repart := `for arg; do case $arg in --root=*) root=${arg#--root=} ;; esac; last=$arg; done
if echo x 2>/dev/null > "$root/etc/miso-null"; then echo opened; else echo refused; fi > "$last"`
	disk := protocol.Disk{Key: "disk", Layers: []string{"base", "boot", "image"}, Tools: fakeTools(t, worker, repart)}

	// act
	err = worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "disk", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, "refused\n", string(written))
}

func TestADiskWhoseToolsFailFailsNamingTheExitCodeAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	disk := protocol.Disk{Key: "disk", Layers: bootable(t, worker), Tools: fakeTools(t, worker, "exit 3")}

	// act
	err := worker.Disk(context.Background(), disk, io.Discard)

	// assert
	assert.ErrorContains(t, err, "exit code 3")
	assert.NoDirExists(t, filepath.Join(layers, "disk"))
}
