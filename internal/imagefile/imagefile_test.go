package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/imagefile"
)

func TestFromOpensAStageOnItsBase(t *testing.T) {
	// arrange
	source := "FROM debian:sid\n"

	// act
	stages := imagefile.Parse(source)

	// assert
	assert.Equal(t, "debian:sid", stages[0].Base)
}

func TestFromWithAsNamesTheStage(t *testing.T) {
	// arrange
	source := "FROM debian:sid AS build\n"

	// act
	stages := imagefile.Parse(source)

	// assert
	assert.Equal(t, "build", stages[0].Name)
}

func TestRunKeepsItsCommandVerbatim(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nRUN echo  'a  b' > /etc/motd\n"

	// act
	stages := imagefile.Parse(source)

	// assert
	assert.Equal(t, []string{"echo  'a  b' > /etc/motd"}, stages[0].Commands)
}

func TestInstructionsBelongToTheStageTheyFollow(t *testing.T) {
	// arrange
	source := "FROM debian:sid AS one\nRUN first\nFROM one AS two\nRUN second\n"

	// act
	stages := imagefile.Parse(source)

	// assert
	assert.Equal(t, []imagefile.Stage{
		{Name: "one", Base: "debian:sid", Commands: []string{"first"}},
		{Name: "two", Base: "one", Commands: []string{"second"}},
	}, stages)
}
