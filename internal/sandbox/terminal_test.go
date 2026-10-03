//go:build vmtest

package sandbox_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

func TestAShellRunsOnATerminal(t *testing.T) {
	// arrange
	root := onBase(t)
	// the terminal echoes what is typed, so only the output has the word whole
	in := strings.NewReader("test -t 0 && echo ter\"\"minal\nexit\n")
	var out bytes.Buffer

	// act
	code, err := sandbox.Shell(context.Background(), root, t.TempDir(), protocol.Shell{}, protocol.Terminal{In: in}, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "terminal\r\n")
}

func TestAShellHasTheSizeAndTypeOfTheTerminalOfTheUser(t *testing.T) {
	// arrange
	root := onBase(t)
	shell := protocol.Shell{Term: "xterm-256color", Rows: 24, Cols: 80}
	in := strings.NewReader("stty size\necho $TERM\nexit\n")
	var out bytes.Buffer

	// act
	_, err := sandbox.Shell(context.Background(), root, t.TempDir(), shell, protocol.Terminal{In: in}, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "24 80\r\n")
	assert.Contains(t, out.String(), "xterm-256color\r\n")
}

func TestTheTypeOfTheTerminalOfTheUserBeatsOneOfTheBuildFile(t *testing.T) {
	// arrange
	root := onBase(t)
	shell := protocol.Shell{Env: []string{"TERM=dumb"}, Term: "xterm-256color"}
	in := strings.NewReader("echo $TERM\nexit\n")
	var out bytes.Buffer

	// act
	_, err := sandbox.Shell(context.Background(), root, t.TempDir(), shell, protocol.Terminal{In: in}, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "xterm-256color\r\n")
}

func TestAShellSeesTheSizeTheTerminalOfTheUserChangesTo(t *testing.T) {
	// arrange
	root := onBase(t)
	resized := make(chan protocol.Resize, 1)
	resized <- protocol.Resize{Rows: 40, Cols: 100}
	// the size arrives while the shell runs, so it waits for it
	in := strings.NewReader("until [ \"$(stty size)\" = \"40 100\" ]; do sleep 0.1; done; echo don\"\"e\nexit\n")
	var out bytes.Buffer
	// a size that never arrives fails the test and does not hang it
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// act
	_, err := sandbox.Shell(ctx, root, t.TempDir(), protocol.Shell{}, protocol.Terminal{In: in, Resized: resized}, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "done\r\n")
}

func TestASizeOfNothingLeavesTheSizeOfTheShellAlone(t *testing.T) {
	// arrange
	root := onBase(t)
	resized := make(chan protocol.Resize, 1)
	resized <- protocol.Resize{}
	// the size has time to arrive while the shell waits
	in := strings.NewReader("sleep 1; stty size\nexit\n")
	var out bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// act
	_, err := sandbox.Shell(ctx, root, t.TempDir(), protocol.Shell{Rows: 24, Cols: 80}, protocol.Terminal{In: in, Resized: resized}, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "24 80\r\n")
}

func TestAShellWithACommandRunsItWithItsWordsAsTheyAre(t *testing.T) {
	// arrange
	root := onBase(t)
	shell := protocol.Shell{Command: []string{"echo", "a  b", "$HOME"}}
	var out bytes.Buffer

	// act
	code, err := sandbox.Shell(context.Background(), root, t.TempDir(), shell, protocol.Terminal{In: strings.NewReader("")}, &out)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "a  b $HOME\r\n", out.String())
}

func TestAShellWithACommandAnswersTheExitCodeOfTheCommand(t *testing.T) {
	// arrange
	root := onBase(t)
	shell := protocol.Shell{Command: []string{"sh", "-c", "exit 3"}}

	// act
	code, err := sandbox.Shell(context.Background(), root, t.TempDir(), shell, protocol.Terminal{In: strings.NewReader("")}, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 3, code)
}

func TestControlCStopsACommandOfAShell(t *testing.T) {
	// arrange
	root := onBase(t)
	shell := protocol.Shell{Command: []string{"sleep", "100"}}
	in, typed := io.Pipe()
	defer func() { _ = typed.Close() }()
	// the command has started by the time the key is pressed
	time.AfterFunc(time.Second, func() { _, _ = typed.Write([]byte{0x03}) })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// act
	code, err := sandbox.Shell(ctx, root, t.TempDir(), shell, protocol.Terminal{In: in}, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 130, code)
}

func TestACommandThatIsNotFoundAnswers127AsAShellDoes(t *testing.T) {
	// arrange
	root := onBase(t)
	shell := protocol.Shell{Command: []string{"-no-such-program"}}

	// act
	code, err := sandbox.Shell(context.Background(), root, t.TempDir(), shell, protocol.Terminal{In: strings.NewReader("")}, io.Discard)

	// assert
	require.NoError(t, err)
	assert.Equal(t, 127, code)
}
