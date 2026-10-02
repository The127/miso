package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/The127/miso/internal/kernel"
	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

// unpacker unpacks the compressed kernel of a bzImage by the magic number its
// stream starts with, which is how the kernel's own tools tell the
// compressors apart.
const unpacker = `set -e
stream=/run/miso/stream/stream
magic=$(od -An -tx1 -N6 "$stream" | tr -d ' \n')
case $magic in
1f8b*) set -- gzip -dc ;;
fd377a585a00) set -- xz -dc ;;
28b52ffd*) set -- zstd -dc ;;
425a68*) set -- bzip2 -dc ;;
5d0000*) set -- xz --format=lzma -dc ;;
894c5a4f*) set -- lzop -dc ;;
02214c18*) set -- lz4 -dc ;;
*) echo "no unpacker for a stream that starts with $magic" >&2; exit 1 ;;
esac
"$@" < "$stream" > ` + outputPath

// elfMagic starts every ELF file.
var elfMagic = []byte("\x7fELF")

// unpack keeps the ELF file a kernel of the image unpacks to, as the layer
// of a key. The tools unpack the stream the header of the kernel points at,
// and the size the kernel itself states checks what they made.
func (a *Agent) unpack(ctx context.Context, request protocol.BootPart, out io.Writer) error {
	return a.madeByTools(request.Key, request.Layers, request.Tools, "kernel", func(dir toolsDir) (toolsRun, error) {
		// links in the image mean places in the image, never in the builder VM
		imageFS := place.Open(dir.image.dir).FS()

		found, err := kernel.Find(imageFS, "")
		if err != nil {
			return nil, err
		}

		stream, unpacked, err := keepStream(imageFS, found.Linux, dir.scratch)
		if err != nil {
			return nil, err
		}

		return func(root string) (int, error) {
			unbindOutput, err := bindOutput(root, dir.output)
			if err != nil {
				return 0, err
			}

			defer unbindOutput()

			unbindStream, err := bind(stream, filepath.Join(root, "run", "miso", "stream"))
			if err != nil {
				return 0, err
			}

			defer unbindStream()

			code, err := sandbox.Run(ctx, root, protocol.Run{Command: unpacker}, out)
			if err != nil || code != 0 {
				return code, err
			}

			return 0, checkUnpacked(filepath.Join(dir.output, outputFile), unpacked)
		}, nil
	})
}

// keepStream writes the compressed kernel of a bzImage of the image into a
// directory of its own, and answers that directory and the size the kernel
// says it unpacks to.
func keepStream(image fs.FS, path, scratch string) (dir string, unpacked uint32, err error) {
	linux := filepath.Join(scratch, "linux")
	if err := copyPart(image, path, linux); err != nil {
		return "", 0, err
	}

	file, err := os.Open(linux)
	if err != nil {
		return "", 0, err
	}

	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return "", 0, err
	}

	found, err := kernel.PayloadOf(file, info.Size())
	if err != nil {
		return "", 0, err
	}

	dir = filepath.Join(scratch, "stream")
	if err := os.Mkdir(dir, 0o700); err != nil {
		return "", 0, err
	}

	into, err := os.OpenFile(filepath.Join(dir, "stream"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", 0, err
	}

	_, err = io.Copy(into, io.NewSectionReader(file, found.Offset, found.Length))

	return dir, found.Unpacked, errors.Join(err, into.Close())
}

// checkUnpacked refuses what the tools made when it is no ELF file or is not
// as large as the kernel says. It runs after the tools, with no cap on what
// they write, because the build file is trusted as in docker build.
func checkUnpacked(path string, unpacked uint32) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	if info.Size() != int64(unpacked) {
		return fmt.Errorf("the unpacked kernel is %d bytes, the kernel says %d", info.Size(), unpacked)
	}

	magic := make([]byte, len(elfMagic))
	if _, err := io.ReadFull(file, magic); err != nil || !bytes.Equal(magic, elfMagic) {
		return errors.New("the unpacked kernel is no ELF file")
	}

	return nil
}
