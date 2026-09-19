package initramfs

import (
	"io"

	"github.com/u-root/u-root/pkg/cpio"
)

// Write writes an initial ramfs whose init is the program given.
func Write(w io.Writer, init []byte) error {
	records := []cpio.Record{
		cpio.StaticRecord(init, cpio.Info{Name: "init", Mode: cpio.S_IFREG | 0o700}),
	}

	archive := cpio.Newc.Writer(w)
	if err := cpio.WriteRecordsAndDirs(archive, records); err != nil {
		return err
	}

	return cpio.WriteTrailer(archive)
}
