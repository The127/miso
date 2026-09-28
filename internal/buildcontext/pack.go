package buildcontext

import (
	"io"
	"strconv"
	"strings"
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
		packed := Entry{Kind: string(found.kind), Path: found.path, Target: found.target}
		// a link has no mode of its own
		if found.mode != "" {
			mode, err := strconv.ParseUint(found.mode, 8, 32)
			if err != nil {
				return err
			}

			packed.Mode = uint32(mode)
		}

		if found.kind != kindFile {
			return visit(packed, strings.NewReader(""))
		}

		file, err := d.root.Open(found.name)
		if err != nil {
			return err
		}

		defer func() { _ = file.Close() }()

		info, err := file.Stat()
		if err != nil {
			return err
		}

		packed.Size = info.Size()

		return visit(packed, file)
	})
}
