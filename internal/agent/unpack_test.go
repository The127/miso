//go:build vmtest

package agent_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/kernel"
	"github.com/The127/miso/internal/protocol"
)

// bzImage is a kernel file of the boot protocol around a compressed stream,
// ending in the size the kernel says it unpacks to.
func bzImage(t *testing.T, stream []byte, unpacked uint32) []byte {
	t.Helper()

	const setupSects = 2

	setup := make([]byte, (setupSects+1)*512)
	setup[0x1f1] = setupSects
	copy(setup[0x202:], "HdrS")
	binary.LittleEndian.PutUint16(setup[0x206:], 0x20f)
	binary.LittleEndian.PutUint32(setup[0x248:], 0)
	binary.LittleEndian.PutUint32(setup[0x24c:], uint32(len(stream)+4)) //nolint:gosec // a few bytes

	return binary.LittleEndian.AppendUint32(append(setup, stream...), unpacked)
}

// gzipped is what gzip makes of some bytes.
func gzipped(t *testing.T, content []byte) []byte {
	t.Helper()

	var packed bytes.Buffer
	writer := gzip.NewWriter(&packed)
	_, err := writer.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	return packed.Bytes()
}

// withBzImage adds an image layer, which the key image names, whose kernel
// is a file.
func withBzImage(t *testing.T, worker *agent.Agent, file []byte) {
	t.Helper()

	encoded := base64.StdEncoding.EncodeToString(file)
	command := "mkdir -p /usr/lib/modules/99.0 && cd /usr/lib/modules/99.0 && touch initrd && echo " + encoded + " | base64 -d > vmlinuz"
	withImage(t, worker, command)
}

func TestAnUnpackedKernelIsTheELFFileItsStreamUnpacksTo(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	elf := []byte("\x7fELF the kernel, as a file")
	withBzImage(t, worker, bzImage(t, gzipped(t, elf), uint32(len(elf)))) //nolint:gosec // a few bytes
	part := protocol.BootPart{Key: "elf", Layers: []string{"base", "image"}, Tools: []string{"base"}, Part: protocol.PartKernel, ELF: true}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "elf", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, elf, written)
}

func TestAnUnpackedKernelNotAsLargeAsTheKernelSaysFailsNamingBothSizesAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	elf := []byte("\x7fELF the kernel, as a file")
	withBzImage(t, worker, bzImage(t, gzipped(t, elf), uint32(len(elf))+1)) //nolint:gosec // a few bytes
	part := protocol.BootPart{Key: "elf", Layers: []string{"base", "image"}, Tools: []string{"base"}, Part: protocol.PartKernel, ELF: true}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	assert.ErrorContains(t, err, "26 bytes, the kernel says 27")
	assert.NoDirExists(t, filepath.Join(layers, "elf"))
}

func TestAnUnpackedKernelThatIsNoELFFileFailsAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	content := []byte("not an elf file, but of the size it says")
	withBzImage(t, worker, bzImage(t, gzipped(t, content), uint32(len(content)))) //nolint:gosec // a few bytes
	part := protocol.BootPart{Key: "elf", Layers: []string{"base", "image"}, Tools: []string{"base"}, Part: protocol.PartKernel, ELF: true}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	assert.ErrorContains(t, err, "no ELF file")
	assert.NoDirExists(t, filepath.Join(layers, "elf"))
}

func TestAStreamNoToolUnpacksFailsSayingWhatItStartsWith(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	withBzImage(t, worker, bzImage(t, []byte("\x01\x02\x03\x04\x05\x06 unknown"), 9))
	part := protocol.BootPart{Key: "elf", Layers: []string{"base", "image"}, Tools: []string{"base"}, Part: protocol.PartKernel, ELF: true}
	var said strings.Builder

	// act
	err := worker.BootPart(context.Background(), part, &said)

	// assert
	assert.ErrorContains(t, err, "exit code 1")
	assert.Contains(t, said.String(), "no unpacker for a stream that starts with 010203040506")
	assert.NoDirExists(t, filepath.Join(layers, "elf"))
}

func TestAKernelFileThatIsNoBzImageFailsToUnpackAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	withBzImage(t, worker, []byte("just some bytes"))
	part := protocol.BootPart{Key: "elf", Layers: []string{"base", "image"}, Tools: []string{"base"}, Part: protocol.PartKernel, ELF: true}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	assert.ErrorIs(t, err, kernel.ErrNotBzImage)
	assert.NoDirExists(t, filepath.Join(layers, "elf"))
}

func TestAnInitrdThatIsAskedToBeUnpackedIsRefusedAndKeepsNoLayer(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	part := protocol.BootPart{Key: "elf", Layers: []string{"base"}, Tools: []string{"base"}, Part: protocol.PartInitrd, ELF: true}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	assert.ErrorContains(t, err, "initrd")
	assert.NoDirExists(t, filepath.Join(layers, "elf"))
}

// unpackedBy is what an unpacked kernel is when its stream starts with a
// magic number, and the tools have a fake of the one unpacker that should
// run, which writes the arguments it is given after an ELF magic number. It
// answers whether that is what the kernel is, given the arguments.
func unpackedBy(t *testing.T, magic []byte, tool, arguments string) string {
	t.Helper()

	layers := t.TempDir()
	worker := mountedBase(t, layers)
	stream := append(slices.Clone(magic), []byte(" the rest of a stream")...)
	command := "cat > /usr/local/bin/" + tool + " <<'EOF'\n#!/bin/sh\nprintf '\\177ELF %s' \"$*\"\nEOF\nchmod 755 /usr/local/bin/" + tool + "\n"
	fake := protocol.Run{Key: "fake", Layers: []string{"base"}, Command: command}
	code, err := worker.Run(context.Background(), fake, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)

	unpacked := len("\x7fELF " + arguments)
	withBzImage(t, worker, bzImage(t, stream, uint32(unpacked))) //nolint:gosec // a few bytes
	part := protocol.BootPart{Key: "elf", Layers: []string{"base", "image"}, Tools: []string{"base", "fake"}, Part: protocol.PartKernel, ELF: true}
	require.NoError(t, worker.BootPart(context.Background(), part, io.Discard))
	written, err := os.ReadFile(filepath.Join(layers, "elf", "disk.raw"))
	require.NoError(t, err)

	return string(written)
}

func TestAnXzStreamIsUnpackedByXz(t *testing.T) {
	assert.Equal(t, "\x7fELF -dc", unpackedBy(t, []byte{0xfd, '7', 'z', 'X', 'Z', 0x00}, "xz", "-dc"))
}

func TestAnLzmaStreamIsUnpackedByXzAsLzma(t *testing.T) {
	assert.Equal(t, "\x7fELF --format=lzma -dc", unpackedBy(t, []byte{0x5d, 0x00, 0x00, 0x80, 0x00, 0x00}, "xz", "--format=lzma -dc"))
}

func TestAZstdStreamIsUnpackedByZstd(t *testing.T) {
	assert.Equal(t, "\x7fELF -dc", unpackedBy(t, []byte{0x28, 0xb5, 0x2f, 0xfd, 0x04, 0x88}, "zstd", "-dc"))
}

func TestABzip2StreamIsUnpackedByBzip2(t *testing.T) {
	assert.Equal(t, "\x7fELF -dc", unpackedBy(t, []byte("BZh91A"), "bzip2", "-dc"))
}

func TestALzopStreamIsUnpackedByLzop(t *testing.T) {
	assert.Equal(t, "\x7fELF -dc", unpackedBy(t, []byte{0x89, 'L', 'Z', 'O', 0x00, 0x0d}, "lzop", "-dc"))
}

func TestALz4StreamIsUnpackedByLz4(t *testing.T) {
	assert.Equal(t, "\x7fELF -dc", unpackedBy(t, []byte{0x02, 0x21, 0x4c, 0x18, 0x00, 0x00}, "lz4", "-dc"))
}

func TestAnUnpackedKernelWithAPathIsTheELFFileOfThatKernelNotOfTheInstalledOne(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	elf := []byte("\x7fELF the built kernel")
	encoded := base64.StdEncoding.EncodeToString(bzImage(t, gzipped(t, elf), uint32(len(elf)))) //nolint:gosec // a few bytes
	withImage(t, worker, "mkdir -p /src && echo "+encoded+" | base64 -d > /src/bzImage")
	part := protocol.BootPart{Key: "elf", Layers: []string{"base", "image"}, Tools: []string{"base"}, Part: protocol.PartKernel, ELF: true, Path: "/src/bzImage"}

	// act
	err := worker.BootPart(context.Background(), part, io.Discard)

	// assert
	require.NoError(t, err)
	written, err := os.ReadFile(filepath.Join(layers, "elf", "disk.raw"))
	require.NoError(t, err)
	assert.Equal(t, elf, written)
}
