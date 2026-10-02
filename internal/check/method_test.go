package check_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/firmware"
	"github.com/The127/miso/internal/qemu"
)

func TestAKernelIsBootedDirectlyAndNoFirmwareIsWritten(t *testing.T) {
	// arrange
	dir := t.TempDir()
	kernel := qemu.Kernel{Image: "/out/vmlinuz", Initramfs: "/out/initrd.img", CommandLine: "root=/dev/vda rw"}
	boot := check.Boot{Kernel: &kernel, Dir: dir}

	// act
	method, err := boot.BootMethod()

	// assert
	require.NoError(t, err)
	assert.Equal(t, kernel, method)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestWithoutAKernelTheFirmwareIsWrittenAndBoots(t *testing.T) {
	// arrange
	dir := t.TempDir()
	boot := check.Boot{Firmware: firmware.Firmware{Code: []byte("code"), Vars: []byte("vars")}, Dir: dir}

	// act
	method, err := boot.BootMethod()

	// assert
	require.NoError(t, err)
	assert.Equal(t, qemu.Firmware{Code: filepath.Join(dir, "code.fd"), Vars: filepath.Join(dir, "vars.fd")}, method)
	code, err := os.ReadFile(filepath.Join(dir, "code.fd"))
	require.NoError(t, err)
	assert.Equal(t, "code", string(code))
}

func TestAKernelBootsOnTheMicrovmBoard(t *testing.T) {
	// arrange
	kernel := qemu.Kernel{Image: "/out/vmlinuz", Initramfs: "/out/initrd.img", CommandLine: "root=/dev/vda rw"}
	boot := check.Boot{Kernel: &kernel}

	// act
	machine := boot.Machine(kernel, 12345)

	// assert
	assert.True(t, machine.Microvm)
	assert.Equal(t, kernel, machine.Boot)
}

func TestAFirmwareBootsOnTheDefaultBoard(t *testing.T) {
	// arrange
	boot := check.Boot{}

	// act
	machine := boot.Machine(qemu.Firmware{Code: "code.fd", Vars: "vars.fd"}, 12345)

	// assert
	assert.False(t, machine.Microvm)
}
