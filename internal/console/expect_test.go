package console_test

import (
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
	serial := strings.NewReader("booting\nlogin: ")

	// act
	before, err := console.New(serial).Expect(regexp.MustCompile(`login: `))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "booting\n", before)
}

func TestAConsoleEndingBeforeThePatternNamesThePattern(t *testing.T) {
	// arrange
	serial := strings.NewReader("booting\nKernel panic")

	// act
	_, err := console.New(serial).Expect(regexp.MustCompile(`login: `))

	// assert
	assert.ErrorIs(t, err, io.EOF)
	assert.EqualError(t, err, `"login: " never came: EOF`)
}

func TestTextAfterAMatchReachesTheNextExpect(t *testing.T) {
	// arrange
	tty := console.New(io.MultiReader(strings.NewReader("log"), strings.NewReader("in: root\n# ")))
	_, err := tty.Expect(regexp.MustCompile(`login: `))
	require.NoError(t, err)

	// act
	before, err := tty.Expect(regexp.MustCompile(`# `))

	// assert
	require.NoError(t, err)
	assert.Equal(t, "root\n", before)
}

func TestAConsoleThatEndedIsNotReadAgain(t *testing.T) {
	// arrange
	tty := console.New(&lastWords{t: t, words: "login: "})
	_, err := tty.Expect(regexp.MustCompile(`login: `))
	require.NoError(t, err)

	// act
	_, err = tty.Expect(regexp.MustCompile(`# `))

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
		w.t.Fatal("the console was read after it ended")
	}

	w.said = true

	return copy(p, w.words), io.EOF
}

func TestAConsoleThatFailsToReadSaysWhy(t *testing.T) {
	// arrange
	broken := errors.New("the serial port broke")
	tty := console.New(iotest.ErrReader(broken))

	// act
	_, err := tty.Expect(regexp.MustCompile(`login: `))

	// assert
	assert.ErrorIs(t, err, broken)
	assert.EqualError(t, err, `"login: " never came: the serial port broke`)
}
