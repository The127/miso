package agent_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/kernel"
)

func TestAnArm64ImageBootsThroughItsArm64SystemdBootAndStub(t *testing.T) {
	// act
	efi := agent.EFIOf("arm64")

	// assert
	assert.Equal(t, agent.EFI{
		Loader:   "usr/lib/systemd/boot/efi/systemd-bootaa64.efi",
		Stub:     "usr/lib/systemd/boot/efi/linuxaa64.efi.stub",
		Fallback: "BOOTAA64.EFI",
	}, efi)
}

func TestTheESPOfAnArm64ImageFallsBackToItsArm64SystemdBoot(t *testing.T) {
	// arrange
	image := fstest.MapFS{"usr/lib/systemd/boot/efi/systemd-bootaa64.efi": {Data: []byte("loader")}}
	esp := filepath.Join(t.TempDir(), "esp")

	// act
	err := agent.MakeESP(image, t.TempDir(), esp, "arm64")

	// assert
	require.NoError(t, err)
	loader, err := os.ReadFile(filepath.Join(esp, "efi", "EFI", "BOOT", "BOOTAA64.EFI"))
	require.NoError(t, err)
	assert.Equal(t, "loader", string(loader))
}

func TestTheUKIOfAnArm64ImageIsBuiltOnItsArm64Stub(t *testing.T) {
	// arrange
	image := fstest.MapFS{
		"boot/vmlinuz":   {Data: []byte("linux")},
		"boot/initrd":    {Data: []byte("initrd")},
		"etc/os-release": {Data: []byte("ID=debian\n")},
		"usr/lib/systemd/boot/efi/linuxaa64.efi.stub": {Data: []byte("stub")},
	}
	parts := t.TempDir()

	// act
	err := agent.CopyParts(image, kernel.Kernel{Linux: "boot/vmlinuz", Initrd: "boot/initrd"}, parts, "arm64")

	// assert
	require.NoError(t, err)
	stub, err := os.ReadFile(filepath.Join(parts, "stub"))
	require.NoError(t, err)
	assert.Equal(t, "stub", string(stub))
}
