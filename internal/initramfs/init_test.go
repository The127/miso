package initramfs_test

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/initramfs"
)

// static is the start of a static program for x86-64, as much as an init
// must be.
func static(t *testing.T) []byte {
	t.Helper()

	return program(t, elf.PT_LOAD)
}

// program is the start of a program for x86-64 with segments of the types
// given.
func program(t *testing.T, segments ...elf.ProgType) []byte {
	t.Helper()

	return programFor(t, elf.EM_X86_64, segments...)
}

// programFor is the start of a program for a machine, all the ELF header and
// its segments of the types given.
func programFor(t *testing.T, machine elf.Machine, segments ...elf.ProgType) []byte {
	t.Helper()

	var program bytes.Buffer
	header := elf.Header64{
		Ident:     [elf.EI_NIDENT]byte{0x7f, 'E', 'L', 'F', byte(elf.ELFCLASS64), byte(elf.ELFDATA2LSB), byte(elf.EV_CURRENT)},
		Type:      uint16(elf.ET_EXEC),
		Machine:   uint16(machine),
		Version:   uint32(elf.EV_CURRENT),
		Phoff:     64,
		Ehsize:    64,
		Phentsize: 56,
		Phnum:     uint16(len(segments)), //nolint:gosec // a test names a handful of segments
	}
	require.NoError(t, binary.Write(&program, binary.LittleEndian, header))
	for _, segment := range segments {
		require.NoError(t, binary.Write(&program, binary.LittleEndian, elf.Prog64{Type: uint32(segment)})) //nolint:gosec // the segment types of debug/elf fit
	}

	return program.Bytes()
}

func TestAnInitThatIsNoProgramIsRefused(t *testing.T) {
	// arrange
	var archive bytes.Buffer

	// act
	err := initramfs.Write(&archive, []byte("the init"), nil)

	// assert
	assert.ErrorContains(t, err, "the init is no program")
}

func TestAnInitThatNeedsADynamicLoaderIsRefusedNamingCgo(t *testing.T) {
	// arrange
	var archive bytes.Buffer

	// act
	err := initramfs.Write(&archive, program(t, elf.PT_LOAD, elf.PT_INTERP), nil)

	// assert
	assert.ErrorContains(t, err, "CGO_ENABLED=0")
}

func TestAnInitForAnotherMachineIsRefusedNamingIt(t *testing.T) {
	// arrange
	var archive bytes.Buffer

	// act
	err := initramfs.Write(&archive, programFor(t, elf.EM_AARCH64, elf.PT_LOAD), nil)

	// assert
	assert.ErrorContains(t, err, "EM_AARCH64")
}
