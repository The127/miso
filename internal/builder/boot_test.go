package builder_test

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/initramfs"
)

func TestTheBootFilesHoldTheKernelsImage(t *testing.T) {
	// arrange
	kernel := builderkernel.Kernel{Image: []byte("the kernel")}

	// act
	boot, err := builder.WriteBoot(t.TempDir(), "amd64", kernel, static(t))

	// assert
	require.NoError(t, err)
	image, err := os.ReadFile(boot.Kernel)
	require.NoError(t, err)
	assert.Equal(t, "the kernel", string(image))
}

func TestTheBootFilesHoldTheInitAndTheKernelsModules(t *testing.T) {
	// arrange
	modules := []initramfs.Module{{Name: "10-vsock.ko", Content: []byte("module")}}
	kernel := builderkernel.Kernel{Image: []byte("the kernel"), Modules: modules}
	init := static(t)
	var want bytes.Buffer
	require.NoError(t, initramfs.Write(&want, "amd64", init, modules))

	// act
	boot, err := builder.WriteBoot(t.TempDir(), "amd64", kernel, init)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(boot.Initramfs)
	require.NoError(t, err)
	assert.Equal(t, want.Bytes(), written)
}

// static is the start of a static program for x86-64, as much as an init
// must be.
func static(t *testing.T) []byte {
	t.Helper()

	var program bytes.Buffer
	header := elf.Header64{
		Ident:     [elf.EI_NIDENT]byte{0x7f, 'E', 'L', 'F', byte(elf.ELFCLASS64), byte(elf.ELFDATA2LSB), byte(elf.EV_CURRENT)},
		Type:      uint16(elf.ET_EXEC),
		Machine:   uint16(elf.EM_X86_64),
		Version:   uint32(elf.EV_CURRENT),
		Phoff:     64,
		Ehsize:    64,
		Phentsize: 56,
		Phnum:     1,
	}
	require.NoError(t, binary.Write(&program, binary.LittleEndian, header))
	require.NoError(t, binary.Write(&program, binary.LittleEndian, elf.Prog64{Type: uint32(elf.PT_LOAD)}))

	return program.Bytes()
}
