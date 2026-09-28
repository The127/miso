package console_test

import (
	"context"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/console"
)

func TestExpectReturnsWhatCameBeforeTheMatch(t *testing.T) {
	// arrange
	screen := strings.NewReader("booting\nlogin: ")

	// act
	before, err := console.New(screen, io.Discard).Expect(t.Context(), regexp.MustCompile(`login: `))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "booting\n", before)
}

func TestAConsoleEndingBeforeThePatternNamesThePattern(t *testing.T) {
	// arrange
	screen := strings.NewReader("booting\nKernel panic")

	// act
	_, err := console.New(screen, io.Discard).Expect(t.Context(), regexp.MustCompile(`login: `))

	// assert
	assert.ErrorIs(t, err, io.EOF)
	assert.EqualError(t, err, `"login: " never came: EOF`)
}

func TestTextAfterAMatchReachesTheNextExpect(t *testing.T) {
	// arrange
	tty := console.New(io.MultiReader(strings.NewReader("log"), strings.NewReader("in: root\n# ")), io.Discard)
	_, err := tty.Expect(t.Context(), regexp.MustCompile(`login: `))
	require.NoError(t, err)

	// act
	before, err := tty.Expect(t.Context(), regexp.MustCompile(`# `))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "root\n", before)
}

func TestAConsoleThatEndedIsNotReadAgain(t *testing.T) {
	// arrange
	tty := console.New(&lastWords{t: t, words: "login: "}, io.Discard)
	_, err := tty.Expect(t.Context(), regexp.MustCompile(`login: `))
	require.NoError(t, err)

	// act
	_, err = tty.Expect(t.Context(), regexp.MustCompile(`# `))

	// assert
	assert.ErrorIs(t, err, io.EOF)
}

// lastWords hands out its words together with the end, as some readers do
// in one call.
type lastWords struct {
	t     *testing.T
	words string
	said  bool
}

func (w *lastWords) Read(p []byte) (int, error) {
	if w.said {
		w.t.Error("the console was read after it ended")

		return 0, io.EOF
	}

	w.said = true

	return copy(p, w.words), io.EOF
}

func TestAConsoleThatFailsToReadSaysWhy(t *testing.T) {
	// arrange
	broken := errors.New("the serial port broke")
	tty := console.New(iotest.ErrReader(broken), io.Discard)

	// act
	_, err := tty.Expect(t.Context(), regexp.MustCompile(`login: `))

	// assert
	assert.ErrorIs(t, err, broken)
	assert.EqualError(t, err, `"login: " never came: the serial port broke`)
}

func TestAPatternThatNeverCameStillReturnsWhatTheConsoleShowedSinceTheLastMatch(t *testing.T) {
	// arrange
	tty := console.New(strings.NewReader("login: root\nKernel panic"), io.Discard)
	_, err := tty.Expect(t.Context(), regexp.MustCompile(`login: `))
	require.NoError(t, err)

	// act
	shown, err := tty.Expect(t.Context(), regexp.MustCompile(`# `))

	// assert
	require.Error(t, err)
	assert.Equal(t, "root\nKernel panic", shown)
}

func TestExpectReturnsWhenItsContextEnds(t *testing.T) {
	// arrange
	silent, speaker := io.Pipe()
	t.Cleanup(func() { _ = speaker.Close() })
	tty := console.New(silent, io.Discard)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	// act
	_, err := tty.Expect(ctx, regexp.MustCompile(`login: `))

	// assert
	assert.ErrorIs(t, err, context.Canceled)
}

func TestAContextThatEndsStillReturnsWhatTheConsoleShowedSinceTheLastMatch(t *testing.T) {
	// arrange
	silent, speaker := io.Pipe()
	t.Cleanup(func() { _ = speaker.Close() })
	go func() { _, _ = speaker.Write([]byte("booting\nstill booting")) }()
	tty := console.New(silent, io.Discard)
	_, err := tty.Expect(t.Context(), regexp.MustCompile(`booting\n`))
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	// act
	shown, err := tty.Expect(ctx, regexp.MustCompile(`login: `))

	// assert
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, "still booting", shown)
}
