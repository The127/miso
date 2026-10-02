package kernel

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// ErrNotBzImage is a kernel file that is no bzImage of the boot protocol.
var ErrNotBzImage = errors.New("not a bzImage")

// ErrNoPayload is a bzImage that does not say where its compressed kernel is.
var ErrNoPayload = errors.New("no payload")

// Stream is where the compressed kernel is in a bzImage, without the size
// the kernel appends to it.
type Stream struct {
	Offset int64
	Length int64

	// the size the stream unpacks to, as the kernel says
	Unpacked uint32
}

// The places of the setup header. The first boot protocol with a payload is
// 2.08, older kernels do not say where it is.
const (
	setupSectsAt    = 0x1f1
	magicAt         = 0x202
	versionAt       = 0x206
	payloadOffsetAt = 0x248
	payloadLengthAt = 0x24c
	headerEnd       = payloadLengthAt + 4
	payloadSince    = 0x208
	sector          = 512

	// a header with no setup sectors means this many
	defaultSetupSects = 4

	// the kernel appends the size it unpacks to
	trailerSize = 4
	magic       = "HdrS"
)

// PayloadOf finds the compressed kernel in a bzImage of a file of a size.
// The file may come from anyone, so every place it names is checked.
func PayloadOf(file io.ReaderAt, size int64) (Stream, error) {
	header := make([]byte, headerEnd)
	if err := readAt(file, header, 0); err != nil {
		return Stream{}, fmt.Errorf("%w: %w", ErrNotBzImage, err)
	}

	if string(header[magicAt:magicAt+len(magic)]) != magic {
		return Stream{}, ErrNotBzImage
	}

	if binary.LittleEndian.Uint16(header[versionAt:]) < payloadSince {
		return Stream{}, fmt.Errorf("%w: the boot protocol is older than 2.08", ErrNoPayload)
	}

	setup := int64(header[setupSectsAt])
	if setup == 0 {
		setup = defaultSetupSects
	}

	offset := (setup+1)*sector + int64(binary.LittleEndian.Uint32(header[payloadOffsetAt:]))
	length := int64(binary.LittleEndian.Uint32(header[payloadLengthAt:]))
	if length < trailerSize {
		return Stream{}, fmt.Errorf("%w: its length is too short for the size it unpacks to", ErrNoPayload)
	}

	if offset+length > size {
		return Stream{}, fmt.Errorf("%w: it lies outside the file", ErrNoPayload)
	}

	trailer := make([]byte, trailerSize)
	if err := readAt(file, trailer, offset+length-trailerSize); err != nil {
		return Stream{}, fmt.Errorf("%w: %w", ErrNoPayload, err)
	}

	return Stream{Offset: offset, Length: length - trailerSize, Unpacked: binary.LittleEndian.Uint32(trailer)}, nil
}

// readAt fills the buffer from a place. A reader may say EOF along with the
// last bytes, which is no failure when the buffer is full.
func readAt(file io.ReaderAt, buffer []byte, at int64) error {
	n, err := file.ReadAt(buffer, at)
	if n == len(buffer) {
		return nil
	}

	return err
}
