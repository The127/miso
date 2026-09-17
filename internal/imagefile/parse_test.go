package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestInstructionsBelongToTheStageTheyFollow(t *testing.T) {
	// arrange
	source := "FROM debian:sid AS one\nRUN first\nFROM one AS two\nRUN second\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Stage{
		{Line: 1, Name: "one", Base: "debian:sid", Instructions: []imagefile.Instruction{imagefile.Run{Line: 2, Command: "first"}}},
		{Line: 3, Name: "two", Base: "one", Instructions: []imagefile.Instruction{imagefile.Run{Line: 4, Command: "second"}}},
	}, stages)
}

func TestAnInstructionBeforeFromIsRejected(t *testing.T) {
	// arrange
	source := "RUN true\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 1: RUN before FROM")
	assert.ErrorIs(t, err, imagefile.ErrBeforeFrom)
}

func TestAnUnknownInstructionIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nBOGUS value\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: unknown instruction BOGUS")
	assert.ErrorIs(t, err, imagefile.ErrUnknownInstruction)
}

func TestCommentLinesAreIgnored(t *testing.T) {
	// arrange
	source := "# the base\nFROM debian:sid\n#RUN disabled\nRUN true\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Stage{
		{Line: 2, Base: "debian:sid", Instructions: []imagefile.Instruction{imagefile.Run{Line: 4, Command: "true"}}},
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

func TestIndentationBeforeAnInstructionIsIgnored(t *testing.T) {
	// arrange
	source := "FROM debian:sid\n    RUN true\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Run{Line: 2, Command: "true"},
	}, stages[0].Instructions)
}

func TestKeywordsMayBeLowercase(t *testing.T) {
	// arrange
	source := "from debian:sid as build\nrun true\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Stage{
		{Line: 1, Name: "build", Base: "debian:sid", Instructions: []imagefile.Instruction{imagefile.Run{Line: 2, Command: "true"}}},
	}, stages)
}

func TestEveryInstructionRemembersTheLineItStartsOn(t *testing.T) {
	// arrange
	source := "# a comment\nFROM debian:sid\nRUN one && \\\n    two\nENV A=1\nCOPY a /b\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Stage{
		{Line: 2, Base: "debian:sid", Instructions: []imagefile.Instruction{
			imagefile.Run{Line: 3, Command: "one &&     two"},
			imagefile.Env{Line: 5, Key: "A", Value: "1"},
			imagefile.Copy{Line: 6, Sources: []string{"a"}, Destination: "/b"},
		}},
	}, stages)
}
