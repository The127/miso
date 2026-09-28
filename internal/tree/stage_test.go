//go:build vmtest

package tree_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/tree"
)

// onto lands everything a copy takes at one path of the image.
func onto(path string) tree.Land {
	return func(string, string, bool) (string, error) { return path, nil }
}

func TestAFileOfAStageLandsAtTheDestination(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "etc"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "etc", "motd"), []byte("hello\n"), 0o600))
	land := onto("/motd")

	// act
	err := tree.Into(stage, place.Open(image), []string{"/etc/motd"}, land)

	// assert
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(image, "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(got))
}

func TestAFileOfAStageKeepsItsOwnerModeAndTimes(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	tool := filepath.Join(stage, "tool")
	require.NoError(t, os.WriteFile(tool, []byte("#!/bin/sh\n"), 0o600))
	require.NoError(t, os.Chown(tool, 1234, 5678))
	require.NoError(t, os.Chmod(tool, os.ModeSetuid|0o755))
	then := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	require.NoError(t, os.Chtimes(tool, then, then))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/tool"}, onto("/tool"))

	// assert
	require.NoError(t, err)
	info, err := os.Stat(filepath.Join(image, "tool"))
	require.NoError(t, err)
	stat, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	assert.Equal(t, uint32(1234), stat.Uid)
	assert.Equal(t, uint32(5678), stat.Gid)
	assert.Equal(t, os.ModeSetuid|0o755, info.Mode())
	assert.True(t, then.Equal(info.ModTime()), "modified %s", info.ModTime())
}

func TestAFileCapabilityOfAStageOutlivesItsOwnerChange(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	ping := filepath.Join(stage, "ping")
	require.NoError(t, os.WriteFile(ping, nil, 0o600))
	require.NoError(t, os.Chown(ping, 1234, 1234))
	require.NoError(t, unix.Lsetxattr(ping, "security.capability", netRaw, 0))
	require.NoError(t, unix.Lsetxattr(ping, "user.note", []byte("file"), 0))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/ping"}, onto("/ping"))

	// assert
	require.NoError(t, err)
	assert.Equal(t, netRaw, xattr(t, filepath.Join(image, "ping"), "security.capability"))
	assert.Equal(t, "file", string(xattr(t, filepath.Join(image, "ping"), "user.note")))
}

// below lands what a copy takes below one directory of the image.
func below(dir string) tree.Land {
	return func(_, path string, _ bool) (string, error) { return filepath.Join(dir, path), nil }
}

func TestADirectoryOfAStageCopiesWhatIsInIt(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "etc", "sub"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "etc", "motd"), []byte("hello\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "etc", "sub", "issue"), []byte("welcome\n"), 0o600))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/etc"}, below("/copied"))

	// assert
	require.NoError(t, err)
	motd, err := os.ReadFile(filepath.Join(image, "copied", "motd"))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(motd))
	issue, err := os.ReadFile(filepath.Join(image, "copied", "sub", "issue"))
	require.NoError(t, err)
	assert.Equal(t, "welcome\n", string(issue))
}

func TestADirectoryOfAStageKeepsItsOwnerModeAndTimesPastItsChildren(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	dir := filepath.Join(stage, "d")
	require.NoError(t, os.Mkdir(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f"), nil, 0o600))
	require.NoError(t, os.Chown(dir, 1234, 5678))
	require.NoError(t, os.Chmod(dir, os.ModeSticky|0o750))
	then := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	require.NoError(t, os.Chtimes(dir, then, then))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/d"}, below("/copied"))

	// assert
	require.NoError(t, err)
	info, err := os.Stat(filepath.Join(image, "copied"))
	require.NoError(t, err)
	stat, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	assert.Equal(t, uint32(1234), stat.Uid)
	assert.Equal(t, uint32(5678), stat.Gid)
	assert.Equal(t, os.ModeDir|os.ModeSticky|0o750, info.Mode())
	assert.True(t, then.Equal(info.ModTime()), "modified %s", info.ModTime())
}

func TestADirectoryOfTheImageKeepsItsOwnWhenAStageIsCopiedIntoIt(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(stage, "d"), 0o750))
	require.NoError(t, os.Chown(filepath.Join(stage, "d"), 1234, 5678))
	require.NoError(t, os.Mkdir(filepath.Join(image, "etc"), 0o700))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/d"}, below("/etc"))

	// assert
	require.NoError(t, err)
	info, err := os.Stat(filepath.Join(image, "etc"))
	require.NoError(t, err)
	stat, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	assert.Equal(t, uint32(0), stat.Uid)
	assert.Equal(t, os.ModeDir|0o700, info.Mode())
}

func TestALinkOfAStageStaysTheLinkItIs(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	link := filepath.Join(stage, "l")
	require.NoError(t, os.Symlink("/etc/target", link))
	require.NoError(t, os.Lchown(link, 1234, 5678))
	then := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	require.NoError(t, unix.Lutimes(link, []unix.Timeval{unix.NsecToTimeval(then.UnixNano()), unix.NsecToTimeval(then.UnixNano())}))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/l"}, onto("/l"))

	// assert
	require.NoError(t, err)
	text, err := os.Readlink(filepath.Join(image, "l"))
	require.NoError(t, err)
	assert.Equal(t, "/etc/target", text)
	info, err := os.Lstat(filepath.Join(image, "l"))
	require.NoError(t, err)
	stat, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	assert.Equal(t, uint32(1234), stat.Uid)
	assert.True(t, then.Equal(info.ModTime()), "modified %s", info.ModTime())
}

func TestTheWholeStageIsCopiedFromItsRoot(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "etc"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "etc", "os-release"), []byte("ID=miso\n"), 0o600))
	require.NoError(t, os.Symlink("usr/bin", filepath.Join(stage, "bin")))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/"}, below("/"))

	// assert
	require.NoError(t, err)
	release, err := os.ReadFile(filepath.Join(image, "etc", "os-release"))
	require.NoError(t, err)
	assert.Equal(t, "ID=miso\n", string(release))
	text, err := os.Readlink(filepath.Join(image, "bin"))
	require.NoError(t, err)
	assert.Equal(t, "usr/bin", text)
}

// named lands each source under its own name in the root of the image.
func named(source, below string, _ bool) (string, error) {
	return filepath.Join("/", filepath.Base(source), below), nil
}

func TestTwoNamesOfOneFileOfAStageStayOneFile(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(stage, "a"), []byte("one\n"), 0o600))
	require.NoError(t, os.Link(filepath.Join(stage, "a"), filepath.Join(stage, "b")))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/a", "/b"}, named)

	// assert
	require.NoError(t, err)
	a, err := os.Stat(filepath.Join(image, "a"))
	require.NoError(t, err)
	b, err := os.Stat(filepath.Join(image, "b"))
	require.NoError(t, err)
	assert.True(t, os.SameFile(a, b), "a and b are two files")
}

func TestAFileWithTwoNamesThatTwoSourcesNameIsStillCopied(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(stage, "a"), []byte("one\n"), 0o600))
	require.NoError(t, os.Link(filepath.Join(stage, "a"), filepath.Join(stage, "b")))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/a", "/a"}, named)

	// assert
	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(image, "a"))
	require.NoError(t, err)
	assert.Equal(t, "one\n", string(got))
}

func TestAFileWithTwoNamesKeepsItsContentWhenALaterSourceTakesItsFirstPlace(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "x"), 0o750))
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "y"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "x", "a"), []byte("x\n"), 0o600))
	require.NoError(t, os.Link(filepath.Join(stage, "x", "a"), filepath.Join(stage, "x", "c")))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "y", "a"), []byte("y\n"), 0o600))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/x/a", "/y/a", "/x/c"}, named)

	// assert
	require.NoError(t, err)
	a, err := os.ReadFile(filepath.Join(image, "a"))
	require.NoError(t, err)
	assert.Equal(t, "y\n", string(a))
	c, err := os.ReadFile(filepath.Join(image, "c"))
	require.NoError(t, err)
	assert.Equal(t, "x\n", string(c))
}

func TestAFileWithTwoNamesKeepsItsContentWhenALaterLinkTakesItsFirstPlace(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "p"), 0o750))
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "q"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "p", "a"), []byte("p\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "q", "a"), []byte("q\n"), 0o600))
	require.NoError(t, os.Link(filepath.Join(stage, "p", "a"), filepath.Join(stage, "p", "c")))
	require.NoError(t, os.Link(filepath.Join(stage, "q", "a"), filepath.Join(stage, "q", "b")))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/q/b", "/p/a", "/q/a", "/p/c"}, named)

	// assert
	require.NoError(t, err)
	c, err := os.ReadFile(filepath.Join(image, "c"))
	require.NoError(t, err)
	assert.Equal(t, "p\n", string(c))
}

func TestAFileWithTwoNamesKeepsItsContentWhenALaterSymlinkTakesItsFirstPlace(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "x"), 0o750))
	require.NoError(t, os.MkdirAll(filepath.Join(stage, "y"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(stage, "x", "a"), []byte("x\n"), 0o600))
	require.NoError(t, os.Link(filepath.Join(stage, "x", "a"), filepath.Join(stage, "x", "c")))
	require.NoError(t, os.Symlink("elsewhere", filepath.Join(stage, "y", "a")))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/x/a", "/y/a", "/x/c"}, named)

	// assert
	require.NoError(t, err)
	c, err := os.ReadFile(filepath.Join(image, "c"))
	require.NoError(t, err)
	assert.Equal(t, "x\n", string(c))
}

func TestDevicesFifosAndSocketsOfAStageAreMadeAsTheyAre(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	require.NoError(t, unix.Mknod(filepath.Join(stage, "null"), unix.S_IFCHR|0o600, int(unix.Mkdev(1, 3)))) //nolint:gosec // the kernel keeps device numbers in 32 bits
	require.NoError(t, unix.Chmod(filepath.Join(stage, "null"), 0o666))
	require.NoError(t, unix.Mkfifo(filepath.Join(stage, "pipe"), 0o640))
	require.NoError(t, unix.Mknod(filepath.Join(stage, "sock"), unix.S_IFSOCK|0o600, 0))

	// act
	err := tree.Into(stage, place.Open(image), []string{"/"}, below("/"))

	// assert
	require.NoError(t, err)
	var null, pipe, sock unix.Stat_t
	require.NoError(t, unix.Lstat(filepath.Join(image, "null"), &null))
	require.NoError(t, unix.Lstat(filepath.Join(image, "pipe"), &pipe))
	require.NoError(t, unix.Lstat(filepath.Join(image, "sock"), &sock))
	assert.Equal(t, uint32(unix.S_IFCHR|0o666), null.Mode)
	assert.Equal(t, unix.Mkdev(1, 3), null.Rdev)
	assert.Equal(t, uint32(unix.S_IFIFO|0o640), pipe.Mode)
	assert.Equal(t, uint32(unix.S_IFSOCK|0o600), sock.Mode)
}

func TestAHoleInAFileOfAStageStaysAHole(t *testing.T) {
	// arrange
	stage := t.TempDir()
	image := t.TempDir()
	sparse, err := os.Create(filepath.Join(stage, "sparse"))
	require.NoError(t, err)
	_, err = sparse.WriteAt([]byte("x"), 64<<20)
	require.NoError(t, err)
	require.NoError(t, sparse.Close())

	// act
	err = tree.Into(stage, place.Open(image), []string{"/sparse"}, onto("/sparse"))

	// assert
	require.NoError(t, err)
	var copied unix.Stat_t
	require.NoError(t, unix.Stat(filepath.Join(image, "sparse"), &copied))
	assert.Equal(t, int64(64<<20+1), copied.Size)
	assert.Less(t, copied.Blocks*512, int64(1<<20), "the hole takes room")
}
