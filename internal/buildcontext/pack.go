package buildcontext

import (
	"io"
	"strconv"
)

// Entry is one thing a copy carries, with nothing in it that a digest does
// not cover.
type Entry struct {
	// "file", "directory" or "link"
	Kind string

	// as seen from the source, "." for the source itself
	Path string

	// the twelve unix bits, none for a link
	Mode uint32

	// of a link, as written
	Target string

	// of a file's content
	Size int64
}

// Pack hands every entry of a source to visit, in the order of its digest,
// a file with its content.
func (d *Dir) Pack(source string, visit func(Entry, io.Reader) error) error {
	return d.walk(source, func(found entry) error {
		file, err := d.root.Open(found.name)
		if err != nil {
			return err
		}

		defer func() { _ = file.Close() }()

		info, err := file.Stat()
		if err != nil {
			return err
		}

		mode, err := strconv.ParseUint(found.mode, 8, 32)
		if err != nil {
			return err
		}

		return visit(Entry{Kind: string(found.kind), Path: found.path, Mode: uint32(mode), Size: info.Size()}, file)
	})
}
