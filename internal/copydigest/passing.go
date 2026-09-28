package copydigest

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

// Passing is the content of an entry on its way, hashed as it is read, so
// that the host and the builder VM sum it the same way. A content may be a
// disk image of many gigabytes, so it is never held whole.
type Passing struct {
	content io.Reader
	hash    hash.Hash
}

// Pass hashes a content as it is read.
func Pass(content io.Reader) *Passing {
	hash := sha256.New()

	return &Passing{content: io.TeeReader(content, hash), hash: hash}
}

func (p *Passing) Read(b []byte) (int, error) {
	return p.content.Read(b)
}

// Sum is the hash of an entry with all of its content. What was not read
// yet is part of it too, and is read now.
func (p *Passing) Sum(entry Entry) (string, error) {
	if _, err := io.Copy(io.Discard, p.content); err != nil {
		return "", err
	}

	return entry.sum(hex.EncodeToString(p.hash.Sum(nil))), nil
}
