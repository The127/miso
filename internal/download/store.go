package download

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Store keeps bytes on the host under the digest of what arrived.
type Store struct {
	dir    string
	client *http.Client
}

// Open takes the directory the bytes live in and the client that fetches
// them. It touches nothing yet.
func Open(dir string, client *http.Client) *Store {
	return &Store{dir: dir, client: client}
}

// Path is where the bytes with a digest live, sha256:<hex> under sha256/.
func (s *Store) Path(digest string) string {
	return filepath.Join(s.blobs(), strings.TrimPrefix(digest, "sha256:"))
}

// keep writes the bytes to a file that gets its name from their digest.
// Until the last byte is in, the file has a temporary name, so a file
// named by a digest is always whole.
func (s *Store) keep(body io.Reader) (string, error) {
	if err := os.MkdirAll(s.blobs(), 0o750); err != nil {
		return "", err
	}

	file, err := os.CreateTemp(s.blobs(), "download-*")
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(file, hash), body); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())

		return "", err
	}

	if err := file.Close(); err != nil {
		return "", err
	}

	digest := "sha256:" + hex.EncodeToString(hash.Sum(nil))

	return digest, os.Rename(file.Name(), s.Path(digest))
}

func (s *Store) blobs() string {
	return filepath.Join(s.dir, "sha256")
}
