package buildcontext_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/buildcontext"
)

func open(t *testing.T, dir string) *buildcontext.Dir {
	t.Helper()

	context, err := buildcontext.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, context.Close()) })

	return context
}

func write(t *testing.T, dir string, name string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}

func mkdir(t *testing.T, dir string, name string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, name), 0o750))
}

func chmod(t *testing.T, dir string, name string, mode os.FileMode) {
	t.Helper()

	require.NoError(t, os.Chmod(filepath.Join(dir, name), mode))
}

func touch(t *testing.T, dir string, name string, when time.Time) {
	t.Helper()

	require.NoError(t, os.Chtimes(filepath.Join(dir, name), when, when))
}

func TestABuildContextThatIsNotThereCannotBeOpened(t *testing.T) {
	// arrange
	missing := filepath.Join(t.TempDir(), "nope")

	// act
	_, err := buildcontext.Open(missing)

	// assert
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestAMissingPathDoesNotExist(t *testing.T) {
	// arrange
	context := open(t, t.TempDir())

	// act
	_, err := context.Digest("nope")

	// assert
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

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

func TestAChangedFileModeChangesTheDigest(t *testing.T) {
	// arrange
	plain := t.TempDir()
	write(t, plain, "run.sh", "#!/bin/sh")
	chmod(t, plain, "run.sh", 0o644)
	executable := t.TempDir()
	write(t, executable, "run.sh", "#!/bin/sh")
	chmod(t, executable, "run.sh", 0o755)

	// act
	plainDigest, plainErr := open(t, plain).Digest("run.sh")
	executableDigest, executableErr := open(t, executable).Digest("run.sh")

	// assert
	require.NoError(t, plainErr)
	require.NoError(t, executableErr)
	assert.NotEqual(t, plainDigest, executableDigest)
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

func TestAChangedFileDeepInATreeChangesTheDigest(t *testing.T) {
	// arrange
	quiet := t.TempDir()
	write(t, quiet, "etc/systemd/system.conf", "LogLevel=info")
	loud := t.TempDir()
	write(t, loud, "etc/systemd/system.conf", "LogLevel=debug")

	// act
	quietDigest, quietErr := open(t, quiet).Digest("etc")
	loudDigest, loudErr := open(t, loud).Digest("etc")

	// assert
	require.NoError(t, quietErr)
	require.NoError(t, loudErr)
	assert.NotEqual(t, quietDigest, loudDigest)
}

func TestARenamedFileChangesTheDigest(t *testing.T) {
	// arrange
	motd := t.TempDir()
	write(t, motd, "etc/motd", "hello")
	issue := t.TempDir()
	write(t, issue, "etc/issue", "hello")

	// act
	motdDigest, motdErr := open(t, motd).Digest("etc")
	issueDigest, issueErr := open(t, issue).Digest("etc")

	// assert
	require.NoError(t, motdErr)
	require.NoError(t, issueErr)
	assert.NotEqual(t, motdDigest, issueDigest)
}

func TestTheSameTreeUnderAnotherNameKeepsTheDigest(t *testing.T) {
	// arrange
	etc := t.TempDir()
	write(t, etc, "etc/motd", "hello")
	config := t.TempDir()
	write(t, config, "config/motd", "hello")

	// act
	etcDigest, etcErr := open(t, etc).Digest("etc")
	configDigest, configErr := open(t, config).Digest("config")

	// assert
	require.NoError(t, etcErr)
	require.NoError(t, configErr)
	assert.Equal(t, etcDigest, configDigest)
}

func TestAnAddedEmptyDirectoryChangesTheDigest(t *testing.T) {
	// arrange
	without := t.TempDir()
	write(t, without, "etc/motd", "hello")
	with := t.TempDir()
	write(t, with, "etc/motd", "hello")
	mkdir(t, with, "etc/cron.d")

	// act
	withoutDigest, withoutErr := open(t, without).Digest("etc")
	withDigest, withErr := open(t, with).Digest("etc")

	// assert
	require.NoError(t, withoutErr)
	require.NoError(t, withErr)
	assert.NotEqual(t, withoutDigest, withDigest)
}

func TestAChangedDirectoryModeChangesTheDigest(t *testing.T) {
	// arrange
	private := t.TempDir()
	mkdir(t, private, "etc/ssl/private")
	chmod(t, private, "etc/ssl/private", 0o700)
	public := t.TempDir()
	mkdir(t, public, "etc/ssl/private")
	chmod(t, public, "etc/ssl/private", 0o755)

	// act
	privateDigest, privateErr := open(t, private).Digest("etc")
	publicDigest, publicErr := open(t, public).Digest("etc")

	// assert
	require.NoError(t, privateErr)
	require.NoError(t, publicErr)
	assert.NotEqual(t, privateDigest, publicDigest)
}
