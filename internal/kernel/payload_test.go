package kernel_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/kernel"
)

// bzImage is a kernel file of the boot protocol: setup sectors that hold the
// header, then the protected mode code that starts with the compressed
// stream and ends in the size it unpacks to.
func bzImage(t *testing.T, setupSects byte, version uint16, offset uint32, stream string, size uint32) []byte {
	t.Helper()

	// a length of 32 bits, which gosec only sees through the parse
	length, err := strconv.ParseUint(strconv.Itoa(len(stream)+4), 10, 32)
	require.NoError(t, err)

	setup := make([]byte, (int(setupSects)+1)*512)
	setup[0x1f1] = setupSects
	copy(setup[0x202:], "HdrS")
	binary.LittleEndian.PutUint16(setup[0x206:], version)
	binary.LittleEndian.PutUint32(setup[0x248:], offset)
	binary.LittleEndian.PutUint32(setup[0x24c:], uint32(length))

	code := bytes.Repeat([]byte{0}, int(offset))
	code = append(code, stream...)
	code = binary.LittleEndian.AppendUint32(code, size)

	return append(setup, code...)
}

func TestThePayloadOfAKernelIsItsCompressedStreamAndTheSizeItUnpacksTo(t *testing.T) {
	// arrange
	file := bzImage(t, 2, 0x20f, 16, "stream", 1234)

	// act
	found, err := kernel.PayloadOf(bytes.NewReader(file), int64(len(file)))

	// assert
	require.NoError(t, err)
	assert.Equal(t, kernel.Stream{Offset: 3*512 + 16, Length: int64(len("stream")), Unpacked: 1234}, found)
}

func TestAKernelWithoutSetupSectorsHasFourOfThem(t *testing.T) {
	// arrange
	file := bzImage(t, 4, 0x20f, 0, "stream", 7)
	file[0x1f1] = 0

	// act
	found, err := kernel.PayloadOf(bytes.NewReader(file), int64(len(file)))

	// assert
	require.NoError(t, err)
	assert.Equal(t, int64(5*512), found.Offset)
}

func TestAFileWithoutTheBootProtocolsMagicIsNoBzImage(t *testing.T) {
	// arrange
	file := bzImage(t, 2, 0x20f, 0, "stream", 7)
	copy(file[0x202:], "NotS")

	// act
	_, err := kernel.PayloadOf(bytes.NewReader(file), int64(len(file)))

	// assert
	assert.ErrorIs(t, err, kernel.ErrNotBzImage)
}

func TestAFileTooShortForTheHeaderIsNoBzImage(t *testing.T) {
	// arrange
	file := []byte("short")

	// act
	_, err := kernel.PayloadOf(bytes.NewReader(file), int64(len(file)))

	// assert
	assert.ErrorIs(t, err, kernel.ErrNotBzImage)
}

func TestAKernelOfAnOlderProtocolHasNoPayloadToFind(t *testing.T) {
	// arrange
	file := bzImage(t, 2, 0x207, 0, "stream", 7)

	// act
	_, err := kernel.PayloadOf(bytes.NewReader(file), int64(len(file)))

	// assert
	assert.ErrorIs(t, err, kernel.ErrNoPayload)
}

func TestAPayloadThatRunsPastTheEndOfTheFileIsRefused(t *testing.T) {
	// arrange
	file := bzImage(t, 2, 0x20f, 0, "stream", 7)
	file = file[:len(file)-3]

	// act
	_, err := kernel.PayloadOf(bytes.NewReader(file), int64(len(file)))

	// assert
	assert.ErrorIs(t, err, kernel.ErrNoPayload)
}

// eofAtTheEnd is a reader that says EOF along with the last bytes of a read
// that reaches the end, as io.ReaderAt allows.
type eofAtTheEnd struct {
	*bytes.Reader
}

func (r eofAtTheEnd) ReadAt(p []byte, off int64) (int, error) {
	n, err := r.Reader.ReadAt(p, off)
	if err == nil && off+int64(n) == r.Size() {
		err = io.EOF
	}

	return n, err
}

func TestAReaderThatSaysEOFWithTheLastBytesStillGivesThePayload(t *testing.T) {
	// arrange
	file := bzImage(t, 2, 0x20f, 16, "the stream", 1234)
	reader := eofAtTheEnd{bytes.NewReader(file)}

	// act
	found, err := kernel.PayloadOf(reader, int64(len(file)))

	// assert
	require.NoError(t, err)
	assert.Equal(t, int64(1234), int64(found.Unpacked))
}

func TestAPayloadTooShortForTheSizeItUnpacksToIsRefused(t *testing.T) {
	// arrange
	file := bzImage(t, 2, 0x20f, 0, "stream", 7)
	binary.LittleEndian.PutUint32(file[0x24c:], 3)

	// act
	_, err := kernel.PayloadOf(bytes.NewReader(file), int64(len(file)))

	// assert
	assert.ErrorIs(t, err, kernel.ErrNoPayload)
	assert.ErrorContains(t, err, "too short")
}

func TestAPayloadThatStartsFarPastTheEndOfTheFileIsRefused(t *testing.T) {
	// arrange
	file := bzImage(t, 2, 0x20f, 0, "stream", 7)
	binary.LittleEndian.PutUint32(file[0x248:], 0xfffffff0)

	// act
	_, err := kernel.PayloadOf(bytes.NewReader(file), int64(len(file)))

	// assert
	assert.ErrorIs(t, err, kernel.ErrNoPayload)
	assert.ErrorContains(t, err, "outside the file")
}

func TestAKernelOfTheFirstProtocolWithAPayloadHasOne(t *testing.T) {
	// arrange
	file := bzImage(t, 2, 0x208, 16, "stream", 1234)

	// act
	found, err := kernel.PayloadOf(bytes.NewReader(file), int64(len(file)))

	// assert
	require.NoError(t, err)
	assert.Equal(t, int64(len("stream")), found.Length)
}
