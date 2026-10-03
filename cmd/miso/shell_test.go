//go:build kvm

package main_test

import (
	"crypto/rand"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestAShellBeforeAStepSeesTheStepsBeforeItAndNotTheStep(t *testing.T) {
	// arrange
	dir := t.TempDir()
	word := rand.Text()
	file := "FROM debian:13\nRUN echo " + word + " > /first\nRUN echo second > /second\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(file), 0o600))
	shell := exec.CommandContext(t.Context(), miso(t), "shell", "--before", "3", dir) //nolint:gosec // the test names the binary
	shell.Stdin = strings.NewReader("cat /first\ntest -e /second || echo no\"\"thing\nexit\n")

	// act
	said, err := shell.CombinedOutput()

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), word+"\r\n")
	assert.Contains(t, string(said), "nothing\r\n")
}

func TestTheExitCodeOfAShellIsTheExitCodeOfMiso(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN true\n"), 0o600))
	shell := exec.CommandContext(t.Context(), miso(t), "shell", "--before", "2", dir) //nolint:gosec // the test names the binary
	shell.Stdin = strings.NewReader("exit 3\n")

	// act
	said, err := shell.CombinedOutput()

	// assert
	exited, isExit := errors.AsType[*exec.ExitError](err)
	require.True(t, isExit, string(said))
	assert.Equal(t, 3, exited.ExitCode())
}

func TestAShellBeforeALineWithNoStepFailsBeforeABoot(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN true\n"), 0o600))

	// act
	said, err := exec.CommandContext(t.Context(), miso(t), "shell", "--before", "9", dir).CombinedOutput() //nolint:gosec // the test names the binary

	// assert
	require.Error(t, err)
	assert.Contains(t, string(said), "line 9: no step at this line")
}

// terminalOf is a terminal made for a test, and answers its two ends. The
// slave is what the program takes as its terminal, and the master is where
// the test types and reads.
func terminalOf(t *testing.T) (master, slave *os.File) {
	t.Helper()

	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|unix.O_NOCTTY, 0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = master.Close() })

	fd := int(master.Fd())
	require.NoError(t, unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0))
	number, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	require.NoError(t, err)
	slave, err = os.OpenFile("/dev/pts/"+strconv.Itoa(number), os.O_RDWR|unix.O_NOCTTY, 0)
	require.NoError(t, err)
	t.Cleanup(func() { _ = slave.Close() })

	return master, slave
}

// shown is what a program printed on a terminal, read as it comes.
type shown struct {
	mu   sync.Mutex
	text strings.Builder
}

func (s *shown) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.text.String()
}

func (s *shown) readFrom(master *os.File) {
	piece := make([]byte, 4096)
	for {
		n, err := master.Read(piece)
		s.mu.Lock()
		s.text.Write(piece[:n])
		s.mu.Unlock()

		if err != nil {
			return
		}
	}
}

// within waits until what was shown holds the text.
func within(t *testing.T, seen *shown, text string) {
	t.Helper()

	assert.Eventually(t, func() bool { return strings.Contains(seen.String(), text) }, time.Minute, 50*time.Millisecond, "never showed %q, but %q", text, seen.String())
}

func TestAShellOnATerminalHasItsSizeAndType(t *testing.T) {
	// arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN true\n"), 0o600))
	master, slave := terminalOf(t)
	require.NoError(t, unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: 24, Col: 80}))
	shell := exec.CommandContext(t.Context(), miso(t), "shell", "--before", "2", dir) //nolint:gosec // the test names the binary
	shell.Env = append(os.Environ(), "TERM=xterm-test")
	shell.Stdin, shell.Stdout, shell.Stderr = slave, slave, slave
	shell.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}
	seen := &shown{}
	go seen.readFrom(master)
	require.NoError(t, shell.Start())
	within(t, seen, "# ")

	// act
	_, err := master.WriteString("stty size; echo $TERM\n")
	require.NoError(t, err)

	// assert
	within(t, seen, "24 80\r\n")
	within(t, seen, "xterm-test\r\n")
	_, err = master.WriteString("exit\n")
	require.NoError(t, err)
	require.NoError(t, shell.Wait())
}

// shellOnATerminal starts a shell before the step on line 2 of a build file
// with a step, on a terminal, and waits for its prompt.
func shellOnATerminal(t *testing.T) (shell *exec.Cmd, master, slave *os.File, seen *shown) {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte("FROM debian:13\nRUN true\n"), 0o600))
	master, slave = terminalOf(t)
	require.NoError(t, unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: 24, Col: 80}))
	shell = exec.CommandContext(t.Context(), miso(t), "shell", "--before", "2", dir) //nolint:gosec // the test names the binary
	shell.Stdin, shell.Stdout, shell.Stderr = slave, slave, slave
	shell.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}
	seen = &shown{}
	go seen.readFrom(master)
	require.NoError(t, shell.Start())
	within(t, seen, "# ")

	return shell, master, slave, seen
}

func TestAShellOnATerminalFollowsItsNewSize(t *testing.T) {
	// arrange
	shell, master, _, seen := shellOnATerminal(t)

	// act
	require.NoError(t, unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Row: 50, Col: 120}))
	_, err := master.WriteString("until [ \"$(stty size)\" = \"50 120\" ]; do sleep 0.1; done; echo don\"\"e\n")
	require.NoError(t, err)

	// assert
	within(t, seen, "done\r\n")
	_, err = master.WriteString("exit\n")
	require.NoError(t, err)
	require.NoError(t, shell.Wait())
}

func TestTheTerminalIsItselfAgainOnceTheShellEnds(t *testing.T) {
	// arrange
	shell, master, slave, seen := shellOnATerminal(t)
	_, err := master.WriteString("exit\n")
	require.NoError(t, err)

	// act
	require.NoError(t, shell.Wait())

	// assert
	state, err := unix.IoctlGetTermios(int(slave.Fd()), unix.TCGETS)
	require.NoError(t, err, seen.String())
	assert.NotZero(t, state.Lflag&unix.ECHO)
	assert.NotZero(t, state.Lflag&unix.ICANON)
}

func TestAShellWithoutALineStandsAtTheEndOfTheBuild(t *testing.T) {
	// arrange
	dir := t.TempDir()
	word := rand.Text()
	file := "FROM debian:13\nRUN echo first > /first\nRUN echo " + word + " > /second\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Imagefile"), []byte(file), 0o600))
	shell := exec.CommandContext(t.Context(), miso(t), "shell", dir) //nolint:gosec // the test names the binary
	shell.Stdin = strings.NewReader("cat /second\nexit\n")

	// act
	said, err := shell.CombinedOutput()

	// assert
	require.NoError(t, err, string(said))
	assert.Contains(t, string(said), word+"\r\n")
}
