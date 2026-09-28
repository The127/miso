package buildcontext

import (
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/The127/miso/internal/copydigest"
)

// Digest says what a COPY of this path would put into an image.
func (d *Dir) Digest(path string) (string, error) {
	var sums []string
	err := d.walk(path, func(found entry) error {
		content := ""
		if found.kind == kindFile {
			var err error
			if content, err = d.content(found.name); err != nil {
				return err
			}
		}

		sums = append(sums, found.sum(content))

		return nil
	})
	if err != nil {
		return "", err
	}

	return copydigest.Of(sums), nil
}

// sum is the hash of the entry, with the hash of a file's content.
func (e entry) sum(content string) string {
	return copydigest.Entry{Kind: string(e.kind), Path: e.path, Mode: e.mode, Target: e.target}.Sum(content)
}

// content streams a file into its hash, a source may be a disk image of
// many gigabytes.
func (d *Dir) content(name string) (string, error) {
	file, err := d.open(name)
	if err != nil {
		return "", err
	}

	defer func() { _ = file.Close() }()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
