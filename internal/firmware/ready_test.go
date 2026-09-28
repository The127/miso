package firmware_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/deb/debtest"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/firmware"
)

// kept writes bytes into a store by hand and answers their digest, as a
// download of them would have left them there.
func kept(t *testing.T, dir string, content []byte) string {
	t.Helper()

	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sha256"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sha256", digest), content, 0o600))

	return "sha256:" + digest
}

func TestTheFirmwareIsReadiedFromThePinnedPackage(t *testing.T) {
	// arrange
	pkg := debtest.Package(t, map[string]string{
		"./usr/share/OVMF/OVMF_CODE_4M.fd": "the code",
		"./usr/share/OVMF/OVMF_VARS_4M.fd": "the vars",
	})
	dir := t.TempDir()
	digest := kept(t, dir, pkg)
	store := download.Open(dir, http.DefaultClient)

	// act
	found, err := firmware.ReadyFrom(context.Background(), store, "https://snapshot.invalid/ovmf.deb", digest)

	// assert
	require.NoError(t, err)
	assert.Equal(t, firmware.Firmware{Code: []byte("the code"), Vars: []byte("the vars")}, found)
}
