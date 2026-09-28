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

	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sha256"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sha256", digest), content, 0o600))

	return "sha256:" + digest
}
