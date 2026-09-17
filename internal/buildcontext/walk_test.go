package buildcontext_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/buildcontext"
)

func TestAMissingPathDoesNotExist(t *testing.T) {
	// arrange
	context := open(t, t.TempDir())

	// act
	_, err := context.Digest("nope")

	// assert
	assert.ErrorIs(t, err, fs.ErrNotExist)
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

func TestAChangedLinkTargetChangesTheDigest(t *testing.T) {
	// arrange
	utc := t.TempDir()
	symlink(t, utc, "etc/localtime", "/usr/share/zoneinfo/UTC")
	berlin := t.TempDir()
	symlink(t, berlin, "etc/localtime", "/usr/share/zoneinfo/Europe/Berlin")

	// act
	utcDigest, utcErr := open(t, utc).Digest("etc")
	berlinDigest, berlinErr := open(t, berlin).Digest("etc")

	// assert
	require.NoError(t, utcErr)
	require.NoError(t, berlinErr)
	assert.NotEqual(t, utcDigest, berlinDigest)
}

func TestADirectoryThatCannotBeReadIsAnError(t *testing.T) {
	// arrange
	if os.Geteuid() == 0 {
		t.Skip("root reads every directory")
	}

	dir := t.TempDir()
	mkdir(t, dir, "etc/ssl/private")
	chmod(t, dir, "etc/ssl/private", 0o000)

	// act
	_, err := open(t, dir).Digest("etc")

	// assert
	assert.ErrorIs(t, err, fs.ErrPermission)
}

func TestALinkAsTheSourceIsNotFollowed(t *testing.T) {
	// arrange
	two := t.TempDir()
	write(t, two, "releases/v2/app", "two")
	symlink(t, two, "current", "releases/v2")
	three := t.TempDir()
	write(t, three, "releases/v2/app", "three")
	symlink(t, three, "current", "releases/v2")

	// act
	twoDigest, twoErr := open(t, two).Digest("current")
	threeDigest, threeErr := open(t, three).Digest("current")

	// assert
	require.NoError(t, twoErr)
	require.NoError(t, threeErr)
	assert.Equal(t, twoDigest, threeDigest)
}

func TestASourceThatLeavesTheContextIsRejected(t *testing.T) {
	// arrange
	dir := t.TempDir()
	write(t, dir, "context/motd", "hello")
	write(t, dir, "secret", "hunter2")

	// act
	_, err := open(t, filepath.Join(dir, "context")).Digest("../secret")

	// assert
	assert.ErrorIs(t, err, buildcontext.ErrOutsideContext)
}

func TestALinkThatLeadsOutOfTheContextIsRejected(t *testing.T) {
	// arrange
	dir := t.TempDir()
	write(t, dir, "secrets/password", "hunter2")
	write(t, dir, "context/motd", "hello")
	symlink(t, dir, "context/out", filepath.Join(dir, "secrets"))

	// act
	_, err := open(t, filepath.Join(dir, "context")).Digest("out/password")

	// assert
	assert.ErrorIs(t, err, buildcontext.ErrThroughLink)
}

func TestASourceSpelledAnotherWayKeepsTheDigest(t *testing.T) {
	// arrange
	dir := t.TempDir()
	write(t, dir, "etc/motd", "hello")
	context := open(t, dir)

	// act
	plain, plainErr := context.Digest("etc")
	dotted, dottedErr := context.Digest("./etc")
	slashed, slashedErr := context.Digest("etc/")

	// assert
	require.NoError(t, plainErr)
	require.NoError(t, dottedErr)
	require.NoError(t, slashedErr)
	assert.Equal(t, plain, dotted)
	assert.Equal(t, plain, slashed)
}

func TestAPipeInTheTreeIsRejected(t *testing.T) {
	// arrange
	dir := t.TempDir()
	write(t, dir, "run/motd", "hello")
	require.NoError(t, syscall.Mkfifo(filepath.Join(dir, "run/initctl"), 0o600))

	// act
	_, err := open(t, dir).Digest("run")

	// assert
	assert.ErrorIs(t, err, buildcontext.ErrSpecialFile)
	assert.ErrorContains(t, err, "run/initctl")
}

func TestASetuidBitChangesTheDigest(t *testing.T) {
	// arrange
	plain := t.TempDir()
	write(t, plain, "sudo", "binary")
	chmod(t, plain, "sudo", 0o755)
	setuid := t.TempDir()
	write(t, setuid, "sudo", "binary")
	chmod(t, setuid, "sudo", 0o755|os.ModeSetuid)

	// act
	plainDigest, plainErr := open(t, plain).Digest("sudo")
	setuidDigest, setuidErr := open(t, setuid).Digest("sudo")

	// assert
	require.NoError(t, plainErr)
	require.NoError(t, setuidErr)
	assert.NotEqual(t, plainDigest, setuidDigest)
}

func TestASetgidBitChangesTheDigest(t *testing.T) {
	// arrange
	plain := t.TempDir()
	mkdir(t, plain, "var/mail")
	chmod(t, plain, "var/mail", 0o775)
	setgid := t.TempDir()
	mkdir(t, setgid, "var/mail")
	chmod(t, setgid, "var/mail", 0o775|os.ModeSetgid)

	// act
	plainDigest, plainErr := open(t, plain).Digest("var")
	setgidDigest, setgidErr := open(t, setgid).Digest("var")

	// assert
	require.NoError(t, plainErr)
	require.NoError(t, setgidErr)
	assert.NotEqual(t, plainDigest, setgidDigest)
}

func TestAStickyBitChangesTheDigest(t *testing.T) {
	// arrange
	plain := t.TempDir()
	mkdir(t, plain, "var/tmp")
	chmod(t, plain, "var/tmp", 0o777)
	sticky := t.TempDir()
	mkdir(t, sticky, "var/tmp")
	chmod(t, sticky, "var/tmp", 0o777|os.ModeSticky)

	// act
	plainDigest, plainErr := open(t, plain).Digest("var")
	stickyDigest, stickyErr := open(t, sticky).Digest("var")

	// assert
	require.NoError(t, plainErr)
	require.NoError(t, stickyErr)
	assert.NotEqual(t, plainDigest, stickyDigest)
}

func TestAMovedDirectoryChangesTheDigest(t *testing.T) {
	// arrange
	nested := t.TempDir()
	write(t, nested, "usr/lib/systemd/system.conf", "LogLevel=info")
	beside := t.TempDir()
	mkdir(t, beside, "usr/lib")
	write(t, beside, "usr/systemd/system.conf", "LogLevel=info")

	// act
	nestedDigest, nestedErr := open(t, nested).Digest("usr")
	besideDigest, besideErr := open(t, beside).Digest("usr")

	// assert
	require.NoError(t, nestedErr)
	require.NoError(t, besideErr)
	assert.NotEqual(t, nestedDigest, besideDigest)
}

func TestTheSameTreeUnderANestedSourceKeepsTheDigest(t *testing.T) {
	// arrange
	nested := t.TempDir()
	write(t, nested, "etc/systemd/system.conf", "LogLevel=info")
	flat := t.TempDir()
	write(t, flat, "conf/system.conf", "LogLevel=info")

	// act
	nestedDigest, nestedErr := open(t, nested).Digest("etc/systemd")
	flatDigest, flatErr := open(t, flat).Digest("conf")

	// assert
	require.NoError(t, nestedErr)
	require.NoError(t, flatErr)
	assert.Equal(t, nestedDigest, flatDigest)
}

func TestTheWholeContextIsASourceLikeAnyDirectory(t *testing.T) {
	// arrange
	whole := t.TempDir()
	write(t, whole, "motd", "hello")
	chmod(t, whole, ".", 0o750)
	part := t.TempDir()
	write(t, part, "sub/motd", "hello")
	chmod(t, part, "sub", 0o750)

	// act
	wholeDigest, wholeErr := open(t, whole).Digest(".")
	partDigest, partErr := open(t, part).Digest("sub")

	// assert
	require.NoError(t, wholeErr)
	require.NoError(t, partErr)
	assert.Equal(t, wholeDigest, partDigest)
}

func TestALinkInATreeIsNotFollowed(t *testing.T) {
	// arrange
	two := t.TempDir()
	write(t, two, "releases/app", "two")
	symlink(t, two, "opt/app", "../releases/app")
	three := t.TempDir()
	write(t, three, "releases/app", "three")
	symlink(t, three, "opt/app", "../releases/app")

	// act
	twoDigest, twoErr := open(t, two).Digest("opt")
	threeDigest, threeErr := open(t, three).Digest("opt")

	// assert
	require.NoError(t, twoErr)
	require.NoError(t, threeErr)
	assert.Equal(t, twoDigest, threeDigest)
}

func TestASourceThroughALinkIsRejected(t *testing.T) {
	// arrange
	dir := t.TempDir()
	write(t, dir, "releases/v2/app", "two")
	symlink(t, dir, "current", "releases/v2")

	// act
	_, err := open(t, dir).Digest("current/app")

	// assert
	assert.ErrorIs(t, err, buildcontext.ErrThroughLink)
	assert.ErrorContains(t, err, "current")
}
