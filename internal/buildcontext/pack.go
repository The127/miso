package buildcontext

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

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
	packed := Entry{Kind: string(found.kind), Path: found.path, Target: found.target}
	// a link has no mode of its own
	if found.mode != "" {
		mode, err := strconv.ParseUint(found.mode, 8, 32)
		if err != nil {
			return "", err
		}

		packed.Mode = uint32(mode)
	}

	if found.kind != kindFile {
		return found.sum(""), visit(packed, strings.NewReader(""))
	}

	file, err := d.open(found.name)
	if err != nil {
		return "", err
	}

	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	packed.Size = info.Size()

	hash := sha256.New()
	if err := visit(packed, io.TeeReader(file, hash)); err != nil {
		return "", err
	}

	// what visit left unread is part of the file too
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return found.sum(hex.EncodeToString(hash.Sum(nil))), nil
}
