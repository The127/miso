package console_test

import (
	"io"
	"regexp"
	"strings"
	"testing"

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
	assert.ErrorContains(t, err, "login: ")
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
