package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestEnvKeepsItsPlaceAmongTheRunLines(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nRUN first\nENV A=1\nRUN second\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Run{Command: "first"},
		imagefile.Env{Key: "A", Value: "1"},
		imagefile.Run{Command: "second"},
	}, stages[0].Instructions)
}

func TestEnvWithoutAValueIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nENV A\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: ENV needs KEY=VALUE")
}

func TestEnvTakesSeveralVariablesOnOneLine(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nENV A=1 B=2\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Env{Key: "A", Value: "1"},
		imagefile.Env{Key: "B", Value: "2"},
	}, stages[0].Instructions)
}

func TestAnEnvBeforeFromIsRejected(t *testing.T) {
	// arrange
	source := "ENV A=1\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 1: ENV before FROM")
}

func TestAnEnvWithoutVariablesIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nENV\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: ENV needs KEY=VALUE")
}

func TestAQuotedEnvValueKeepsItsSpaces(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nENV MOTD=\"hello world\" B=2\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Env{Key: "MOTD", Value: "hello world"},
		imagefile.Env{Key: "B", Value: "2"},
	}, stages[0].Instructions)
}

func TestAnUnclosedQuoteInEnvIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nENV A=\"oops\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: ENV has an unclosed quote")
}
