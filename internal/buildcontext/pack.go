package buildcontext

import (
	"fmt"
	"io"

	"github.com/The127/miso/internal/copydigest"
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
// a file with its content. It fails with ErrChanged once the source turns
// out not to be what the digest says, so a key never stands for other bytes
// than the ones planned.
func (d *Dir) Pack(source string, digest string, visit func(Entry, io.Reader) error) error {
	var sums []string
	err := d.walk(source, func(found entry) error {
		sum, err := d.pack(found, visit)
		sums = append(sums, sum)

		return err
	})
	if err != nil {
		return err
	}

	if copydigest.Of(sums) != digest {
		return fmt.Errorf("%s: %w", source, ErrChanged)
	}

	return nil
}

// pack hands one entry to visit and sums it as it went out.
func (d *Dir) pack(found entry, visit func(Entry, io.Reader) error) (string, error) {
	content, size, err := d.contentOf(found)
	if err != nil {
		return "", err
	}

	defer func() { _ = content.Close() }()

	passing := copydigest.Pass(content)
	packed := Entry{Kind: string(found.kind), Path: found.path, Mode: found.mode, Target: found.target, Size: size}
	if err := visit(packed, passing); err != nil {
		return "", err
	}

	return passing.Sum(found.digested())
}
