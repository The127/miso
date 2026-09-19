package initramfs_test

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/u-root/u-root/pkg/cpio"

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

// record is the file of a name in an archive, which must hold it.
func record(t *testing.T, archive []byte, name string) cpio.Record {
	t.Helper()

	records, err := cpio.ReadAllRecords(cpio.Newc.Reader(bytes.NewReader(archive)))
	require.NoError(t, err)
	for _, found := range records {
		if found.Name == name {
			return found
		}
	}

	require.Failf(t, "not in the archive", "no %s in the archive", name)

	return cpio.Record{}
}

func TestAnInitramfsHoldsItsInitAsAnExecutableInit(t *testing.T) {
	// arrange
	var archive bytes.Buffer

	// act
	err := initramfs.Write(&archive, static(t), nil)

	// assert
	require.NoError(t, err)
	init := record(t, archive.Bytes(), "init")
	assert.Equal(t, uint64(cpio.S_IFREG|0o700), init.Mode)
	content := make([]byte, init.FileSize)
	_, err = init.ReadAt(content, 0)
	require.NoError(t, err)
	assert.Equal(t, static(t), content)
}

func TestAnInitramfsHoldsTheConsole(t *testing.T) {
	// arrange
	var archive bytes.Buffer

	// act
	err := initramfs.Write(&archive, static(t), nil)

	// assert
	require.NoError(t, err)
	console := record(t, archive.Bytes(), "dev/console")
	assert.Equal(t, uint64(cpio.S_IFCHR|0o600), console.Mode)
	assert.Equal(t, uint64(5), console.Rmajor)
	assert.Equal(t, uint64(1), console.Rminor)
}

func TestAnInitramfsHoldsItsModulesInTheOrderToLoadThem(t *testing.T) {
	// arrange
	var archive bytes.Buffer
	modules := []initramfs.Module{
		{Name: "virtio_blk", Content: []byte("first")},
		{Name: "btrfs", Content: []byte("second")},
	}

	// act
	err := initramfs.Write(&archive, static(t), modules)

	// assert
	require.NoError(t, err)
	first := record(t, archive.Bytes(), "modules/01-virtio_blk.ko")
	second := record(t, archive.Bytes(), "modules/02-btrfs.ko")
	assert.Equal(t, uint64(5), first.FileSize)
	assert.Equal(t, uint64(6), second.FileSize)
}

func TestAnInitramfsCarriesNothingOfTheMachineThatWroteIt(t *testing.T) {
	// arrange
	var archive bytes.Buffer
	modules := []initramfs.Module{{Name: "virtio_blk", Content: []byte("first")}}

	// act
	err := initramfs.Write(&archive, static(t), modules)

	// assert
	require.NoError(t, err)
	records, err := cpio.ReadAllRecords(cpio.Newc.Reader(bytes.NewReader(archive.Bytes())))
	require.NoError(t, err)
	for _, found := range records {
		assert.Zero(t, found.MTime, found.Name)
		assert.Zero(t, found.UID, found.Name)
		assert.Zero(t, found.GID, found.Name)
		assert.Zero(t, found.Major, found.Name)
		assert.Zero(t, found.Minor, found.Name)
	}
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
