package buildcontext_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAChangedFileContentChangesTheDigest(t *testing.T) {
	// arrange
	hello := t.TempDir()
	write(t, hello, "motd", "hello")
	goodbye := t.TempDir()
	write(t, goodbye, "motd", "goodbye")

	// act
	helloDigest, helloErr := open(t, hello).Digest("motd")
	goodbyeDigest, goodbyeErr := open(t, goodbye).Digest("motd")

	// assert
	require.NoError(t, helloErr)
	require.NoError(t, goodbyeErr)
	assert.NotEqual(t, helloDigest, goodbyeDigest)
}

func TestAFileThatCannotBeReadIsAnError(t *testing.T) {
	// arrange
	if os.Geteuid() == 0 {
		t.Skip("root reads every file")
	}

	dir := t.TempDir()
	write(t, dir, "shadow", "secret")
	chmod(t, dir, "shadow", 0o000)

	// act
	_, err := open(t, dir).Digest("shadow")

	// assert
	assert.ErrorIs(t, err, fs.ErrPermission)
}

func TestAChangedModificationTimeKeepsTheDigest(t *testing.T) {
	// arrange
	old := t.TempDir()
	write(t, old, "motd", "hello")
	touch(t, old, "motd", time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC))
	recent := t.TempDir()
	write(t, recent, "motd", "hello")
	touch(t, recent, "motd", time.Date(2026, time.September, 17, 0, 0, 0, 0, time.UTC))

	// act
	oldDigest, oldErr := open(t, old).Digest("motd")
	recentDigest, recentErr := open(t, recent).Digest("motd")

	// assert
	require.NoError(t, oldErr)
	require.NoError(t, recentErr)
	assert.Equal(t, oldDigest, recentDigest)
}

func TestALinkIsNotAFileThatLooksLikeIt(t *testing.T) {
	// arrange
	sum := sha256.Sum256([]byte("hello"))
	file := t.TempDir()
	write(t, file, "etc/motd", "hello")
	chmod(t, file, "etc/motd", 0o777)
	link := t.TempDir()
	symlink(t, link, "etc/motd", hex.EncodeToString(sum[:]))

	// act
	fileDigest, fileErr := open(t, file).Digest("etc")
	linkDigest, linkErr := open(t, link).Digest("etc")

	// assert
	require.NoError(t, fileErr)
	require.NoError(t, linkErr)
	assert.NotEqual(t, fileDigest, linkDigest)
}

func TestANameDoesNotRunIntoTheMode(t *testing.T) {
	// arrange
	setuid := t.TempDir()
	write(t, setuid, "bin/sudo", "binary")
	chmod(t, setuid, "bin/sudo", 0o755|os.ModeSetuid)
	numbered := t.TempDir()
	write(t, numbered, "bin/sudo4", "binary")
	chmod(t, numbered, "bin/sudo4", 0o755)

	// act
	setuidDigest, setuidErr := open(t, setuid).Digest("bin")
	numberedDigest, numberedErr := open(t, numbered).Digest("bin")

	// assert
	require.NoError(t, setuidErr)
	require.NoError(t, numberedErr)
	assert.NotEqual(t, setuidDigest, numberedDigest)
}

func TestADigestNamesItsFormat(t *testing.T) {
	// arrange
	dir := t.TempDir()
	write(t, dir, "motd", "hello")

	// act
	digest, err := open(t, dir).Digest("motd")

	// assert
	require.NoError(t, err)
	assert.Regexp(t, `^miso-context-1:[0-9a-f]{64}$`, digest)
}
