package initramfs

import (
	"fmt"
	"io"

	"github.com/u-root/u-root/pkg/cpio"
)

// Write writes an initial ramfs whose init is the program given, with the
// modules in the order to load them.
func Write(w io.Writer, init []byte, modules []Module) error {
	records := []cpio.Record{
		cpio.StaticRecord(init, cpio.Info{Name: "init", Mode: cpio.S_IFREG | 0o700}),
		// init has no output without it, and only some kernels bring one
		cpio.CharDev("dev/console", 0o600, 5, 1),
	}

	// the agent loads them in the order of their names
	for i, module := range modules {
		name := fmt.Sprintf("modules/%02d-%s.ko", i+1, module.Name)
		records = append(records, cpio.StaticRecord(module.Content, cpio.Info{Name: name, Mode: cpio.S_IFREG | 0o600}))
	}

	archive := cpio.Newc.Writer(w)
	if err := cpio.WriteRecordsAndDirs(archive, records); err != nil {
		return err
	}

	return cpio.WriteTrailer(archive)
}
