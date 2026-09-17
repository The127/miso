package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestFromOpensAStageOnItsBase(t *testing.T) {
	// arrange
	source := "FROM debian:sid\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "debian:sid", stages[0].Base)
}

func TestFromWithAsNamesTheStage(t *testing.T) {
	// arrange
	source := "FROM debian:sid AS build\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "build", stages[0].Name)
}

func TestRunKeepsItsCommandVerbatim(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nRUN echo  'a  b' > /etc/motd\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Run{Command: "echo  'a  b' > /etc/motd"},
	}, stages[0].Instructions)
}

func TestInstructionsBelongToTheStageTheyFollow(t *testing.T) {
	// arrange
	source := "FROM debian:sid AS one\nRUN first\nFROM one AS two\nRUN second\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Stage{
		{Name: "one", Base: "debian:sid", Instructions: []imagefile.Instruction{imagefile.Run{Command: "first"}}},
		{Name: "two", Base: "one", Instructions: []imagefile.Instruction{imagefile.Run{Command: "second"}}},
	}, stages)
}

func TestAnInstructionBeforeFromIsRejected(t *testing.T) {
	// arrange
	source := "RUN true\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 1: RUN before FROM")
}

func TestAnUnknownInstructionIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nBOGUS value\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: unknown instruction BOGUS")
}

func TestCommentLinesAreIgnored(t *testing.T) {
	// arrange
	source := "# the base\nFROM debian:sid\n#RUN disabled\nRUN true\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Stage{
		{Base: "debian:sid", Instructions: []imagefile.Instruction{imagefile.Run{Command: "true"}}},
	}, stages)
}

func TestAFileWithoutFromIsRejected(t *testing.T) {
	// arrange
	source := "# nothing but a comment\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "no FROM instruction")
}

func TestFromWithoutABaseIsRejected(t *testing.T) {
	// arrange
	source := "FROM\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 1: FROM needs a base")
}

func TestFromWithStrayWordsIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid build\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 1: FROM takes a base and an optional AS name")
}

func TestABackslashContinuesTheInstructionOnTheNextLine(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nRUN apt-get update && \\\n    apt-get install -y vim\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Run{Command: "apt-get update &&     apt-get install -y vim"},
	}, stages[0].Instructions)
}

func TestLineNumbersStayTrueAfterAContinuation(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nRUN true && \\\n    true\nBOGUS\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 4: unknown instruction BOGUS")
}

func TestAContinuationThatRunsOffTheEndOfTheFileIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nRUN one && \\\n    two && \\\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: continuation runs off the end of the file")
}

func TestIndentationBeforeAnInstructionIsIgnored(t *testing.T) {
	// arrange
	source := "FROM debian:sid\n    RUN true\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Run{Command: "true"},
	}, stages[0].Instructions)
}

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
