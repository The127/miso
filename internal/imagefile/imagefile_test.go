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
	assert.Equal(t, []string{"echo  'a  b' > /etc/motd"}, stages[0].Commands)
}

func TestInstructionsBelongToTheStageTheyFollow(t *testing.T) {
	// arrange
	source := "FROM debian:sid AS one\nRUN first\nFROM one AS two\nRUN second\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Stage{
		{Name: "one", Base: "debian:sid", Commands: []string{"first"}},
		{Name: "two", Base: "one", Commands: []string{"second"}},
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
		{Base: "debian:sid", Commands: []string{"true"}},
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
