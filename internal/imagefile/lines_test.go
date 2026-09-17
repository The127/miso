package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

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

func TestAContinuationOntoAnEmptyLineIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nRUN one && \\\n\nRUN two\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: continuation onto an empty line")
}
