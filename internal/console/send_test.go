package console_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/console"
)

func TestSendTypesTheTextOnTheConsole(t *testing.T) {
	// arrange
	var keyboard bytes.Buffer
	tty := console.New(strings.NewReader(""), &keyboard)

	// act
	err := tty.Send("root\n")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "root\n", keyboard.String())
}

func TestAConsoleThatFailsToTypeSaysWhy(t *testing.T) {
	// arrange
	unplugged := errors.New("the keyboard is unplugged")
	tty := console.New(strings.NewReader(""), brokenKeyboard{unplugged})

	// act
	err := tty.Send("root\n")

	// assert
	assert.ErrorIs(t, err, unplugged)
	assert.EqualError(t, err, `typing on the console: the keyboard is unplugged`)
}

type brokenKeyboard struct {
	err error
}

func (k brokenKeyboard) Write([]byte) (int, error) {
	return 0, k.err
}
