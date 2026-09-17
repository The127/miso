package imagefile_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
)

func TestCopyNamesItsSourcesAndItsDestination(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCOPY go.mod go.sum /src/\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Copy{Line: 2, Sources: []string{"go.mod", "go.sum"}, Destination: "/src/"},
	}, stages[0].Instructions)
}

func TestCopyWithoutADestinationIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCOPY onlyone\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: COPY needs a source and a destination")
}

func TestCopyFromNamesTheStageItCopiesFrom(t *testing.T) {
	// arrange
	source := "FROM debian:sid AS build\nFROM scratch\nCOPY --from=build /out/app /usr/bin/\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Copy{Line: 3, From: "build", Sources: []string{"/out/app"}, Destination: "/usr/bin/"},
	}, stages[1].Instructions)
}

func TestCopyWithAnUnknownFlagIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCOPY --frm=build a b\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: COPY does not know --frm")
}

func TestCopyFromWithoutAStageIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCOPY --from= a b\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: COPY --from needs a stage")
}

func TestAQuotedCopyPathKeepsItsSpaces(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCOPY \"my file.txt\" /srv/\n"

	// act
	stages, err := imagefile.Parse(source)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []imagefile.Instruction{
		imagefile.Copy{Line: 2, Sources: []string{"my file.txt"}, Destination: "/srv/"},
	}, stages[0].Instructions)
}

func TestAnUnclosedQuoteInCopyIsRejected(t *testing.T) {
	// arrange
	source := "FROM debian:sid\nCOPY \"oops /srv/\n"

	// act
	_, err := imagefile.Parse(source)

	// assert
	assert.EqualError(t, err, "line 2: COPY has an unclosed quote")
}
