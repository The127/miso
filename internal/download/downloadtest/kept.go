package downloadtest

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Kept writes bytes into the store in dir by hand and answers their digest.
func Kept(t *testing.T, dir string, content []byte) string {
	t.Helper()

	digest := Digest(content)
	KeptAs(t, dir, digest, content)

	return digest
}

// KeptAs writes bytes into the store in dir by hand under a digest, which
// need not be theirs.
func KeptAs(t *testing.T, dir, digest string, content []byte) {
	t.Helper()

	path := filepath.Join(dir, "sha256", digest[len("sha256:"):])
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, content, 0o600))
}

// Digest is the digest the store names bytes by.
func Digest(content []byte) string {
	sum := sha256.Sum256(content)

	return "sha256:" + hex.EncodeToString(sum[:])
}
