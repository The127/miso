package builderkernel_test

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builderkernel"
)

// module is a kernel module: an ELF file whose .modinfo section holds the
// entries, as the kernel's own modules do.
func module(t *testing.T, entries ...string) []byte {
	t.Helper()

	modinfo := []byte(strings.Join(entries, "\x00") + "\x00")
	names := []byte("\x00.modinfo\x00.shstrtab\x00")
	header := elf.Header64{
		Ident:     [elf.EI_NIDENT]byte{0x7f, 'E', 'L', 'F', byte(elf.ELFCLASS64), byte(elf.ELFDATA2LSB), byte(elf.EV_CURRENT)},
		Type:      uint16(elf.ET_REL),
		Machine:   uint16(elf.EM_X86_64),
		Version:   uint32(elf.EV_CURRENT),
		Ehsize:    64,
		Shentsize: 64,
		Shnum:     3,
		Shstrndx:  2,
		Shoff:     uint64(64 + len(modinfo) + len(names)),
	}

	sections := []elf.Section64{
		{},
		{Name: 1, Type: uint32(elf.SHT_PROGBITS), Off: 64, Size: uint64(len(modinfo))},
		{Name: 10, Type: uint32(elf.SHT_STRTAB), Off: uint64(64 + len(modinfo)), Size: uint64(len(names))},
	}

	var file bytes.Buffer
	require.NoError(t, binary.Write(&file, binary.LittleEndian, header))
	file.Write(modinfo)
	file.Write(names)
	for _, section := range sections {
		require.NoError(t, binary.Write(&file, binary.LittleEndian, section))
	}

	return file.Bytes()
}

// withoutSections is an ELF file that holds no sections at all.
func withoutSections(t *testing.T) []byte {
	t.Helper()

	header := elf.Header64{
		Ident:     [elf.EI_NIDENT]byte{0x7f, 'E', 'L', 'F', byte(elf.ELFCLASS64), byte(elf.ELFDATA2LSB), byte(elf.EV_CURRENT)},
		Type:      uint16(elf.ET_REL),
		Machine:   uint16(elf.EM_X86_64),
		Version:   uint32(elf.EV_CURRENT),
		Ehsize:    64,
		Shentsize: 64,
	}

	var file bytes.Buffer
	require.NoError(t, binary.Write(&file, binary.LittleEndian, header))

	return file.Bytes()
}

func TestAModuleNamesItselfAndWhatItNeeds(t *testing.T) {
	// arrange
	ko := module(t, "license=GPL", "depends=xor,raid6_pq,libcrc32c", "name=btrfs")

	// act
	info, err := builderkernel.Modinfo(bytes.NewReader(ko))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "btrfs", info.Name)
	assert.Equal(t, []string{"xor", "raid6_pq", "libcrc32c"}, info.Depends)
}

func TestSomethingThatIsNoModuleIsRefused(t *testing.T) {
	// arrange
	notAModule := withoutSections(t)

	// act
	_, err := builderkernel.Modinfo(bytes.NewReader(notAModule))

	// assert
	assert.ErrorContains(t, err, "no .modinfo")
}
