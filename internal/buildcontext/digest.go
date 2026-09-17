package buildcontext

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strconv"
)

// format names the way a digest is made. A change to what goes into a
// digest, or how, gets a new name, so that no old digest can match a new one.
const format = "miso-context-1"

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

	return format + ":" + hashed(sums), nil
}

// sum takes the hash of a file's content, so that a payload can stream the
// content first and sum the entry after.
func (e entry) sum(content string) string {
	payload := e.target
	if e.kind == kindFile {
		payload = content
	}

	return hashed([]string{string(e.kind), e.path, e.mode, payload})
}

// content streams a file into its hash, a source may be a disk image of
// many gigabytes.
func (d *Dir) content(name string) (string, error) {
	file, err := d.root.Open(name)
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

// hashed puts the length in front of every field, so that no field can
// run into the next.
func hashed(fields []string) string {
	hash := sha256.New()
	for _, field := range fields {
		hash.Write([]byte(strconv.Itoa(len(field)) + ":" + field))
	}

	return hex.EncodeToString(hash.Sum(nil))
}
