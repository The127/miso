package emulator_test

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/emulator"
)

func TestTheEmulatorOfAnArm64ImageIsTheHostsStaticQEMUOnThePath(t *testing.T) {
	// arrange
	dir := t.TempDir()
	program := program(t, uint32(elf.PT_LOAD))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "qemu-aarch64-static"), program, 0o600))
	t.Setenv("PATH", dir)

	// act
	found, err := emulator.Host("arm64")

	// assert
	require.NoError(t, err)
	assert.Equal(t, program, found)
}

func TestTheEmulatorOfAnAmd64ImageIsTheHostsStaticQEMUOnThePath(t *testing.T) {
	// arrange
	dir := t.TempDir()
	program := program(t, uint32(elf.PT_LOAD))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "qemu-aarch64-static"), []byte("another emulator"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "qemu-x86_64-static"), program, 0o600))
	t.Setenv("PATH", dir)

	// act
	found, err := emulator.Host("amd64")

	// assert
	require.NoError(t, err)
	assert.Equal(t, program, found)
}

func TestAnEmulatorThatIsNotInstalledIsRefusedNamingThePackageItComesWith(t *testing.T) {
	// arrange
	t.Setenv("PATH", t.TempDir())

	// act
	_, err := emulator.Host("amd64")

	// assert
	require.ErrorContains(t, err, "amd64")
	assert.ErrorContains(t, err, "qemu-x86_64-static")
	assert.ErrorContains(t, err, "qemu-user-static")
}

func TestAnEmulatorThatIsNotBuiltStaticIsRefusedNamingIt(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "qemu-x86_64-static"), program(t, uint32(elf.PT_INTERP)), 0o600))
	t.Setenv("PATH", dir)

	// act
	_, err := emulator.Host("amd64")

	// assert
	require.ErrorContains(t, err, filepath.Join(dir, "qemu-x86_64-static"))
	assert.ErrorContains(t, err, "static")
}

func TestAnEmulatorOfAnArchitectureMisoHasNoQEMUForIsRefusedNamingIt(t *testing.T) {
	// arrange
	t.Setenv("PATH", t.TempDir())

	// act
	_, err := emulator.Host("sparc")

	// assert
	require.ErrorContains(t, err, "sparc")
	assert.NotContains(t, err.Error(), "qemu-user-static")
}

// program is an ELF file with one program header of a type.
func program(t *testing.T, kind uint32) []byte {
	t.Helper()

	var file bytes.Buffer
	header := elf.Header64{
		Ident:     [elf.EI_NIDENT]byte{0x7f, 'E', 'L', 'F', byte(elf.ELFCLASS64), byte(elf.ELFDATA2LSB), byte(elf.EV_CURRENT)},
		Type:      uint16(elf.ET_EXEC),
		Machine:   uint16(elf.EM_AARCH64),
		Version:   uint32(elf.EV_CURRENT),
		Phoff:     64,
		Ehsize:    64,
		Phentsize: 56,
		Phnum:     1,
		Shentsize: 64,
	}
	require.NoError(t, binary.Write(&file, binary.LittleEndian, header))
	require.NoError(t, binary.Write(&file, binary.LittleEndian, elf.Prog64{Type: kind}))

	return file.Bytes()
}
