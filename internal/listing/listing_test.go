package listing_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/listing"
	"github.com/The127/miso/internal/plan"
)

func TestAStepShowsItsLineItsKeyAndItsCommand(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "apt-get install -y vim"}, Key: "9a8b7c6d5e4f"}},
	}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM scratch\n   2  9a8b7c6d5e4f  RUN apt-get install -y vim\n", out.String())
}

func TestAKeyIsShortenedToTwelveCharacters(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "true"}, Key: "9a8b7c6d5e4f0011223344556677889900aabbccddeeff00112233445566778899"}},
	}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM scratch\n   2  9a8b7c6d5e4f  RUN true\n", out.String())
}

func TestANamedStageShowsItsName(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{Name: "rootfs", Base: "scratch"}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM scratch AS rootfs\n", out.String())
}

func TestAStageOnAnImageShowsTheDigestOfItsBase(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{Base: "debian:sid", BaseDigest: "47348ce3c15ba0348ac0887f85dd16b27501e538ff66fc2756c2fa642dc4102c"}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM debian:sid  base 47348ce3c15b\n", out.String())
}

func TestAContextCopyListsItsFilesWithTheirDigests(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base: "scratch",
		Steps: []plan.Step{{
			Instruction: imagefile.Copy{Line: 2, Sources: []string{"etc/motd", "etc/issue"}, Destination: "/etc/"},
			Key:         "9a8b7c6d5e4f",
			Files: []plan.File{
				{Path: "etc/motd", Digest: "47348ce3c15ba0348ac0887f85dd16b27501e538ff66fc2756c2fa642dc4102c"},
				{Path: "etc/issue", Digest: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
			},
		}},
	}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM scratch\n"+
		"   2  9a8b7c6d5e4f  COPY etc/motd etc/issue /etc/\n"+
		"                    etc/motd  47348ce3c15b\n"+
		"                    etc/issue  2cf24dba5fb0\n", out.String())
}

func TestADigestIsShortenedAfterItsFormatName(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base: "scratch",
		Steps: []plan.Step{{
			Instruction: imagefile.Copy{Line: 2, Sources: []string{"etc/motd"}, Destination: "/etc/"},
			Key:         "9a8b7c6d5e4f",
			Files:       []plan.File{{Path: "etc/motd", Digest: "miso-context-1:47348ce3c15ba0348ac0887f85dd16b27501e538ff66fc2756c2fa642dc4102c"}},
		}},
	}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM scratch\n"+
		"   2  9a8b7c6d5e4f  COPY etc/motd /etc/\n"+
		"                    etc/motd  miso-context-1:47348ce3c15b\n", out.String())
}

func TestAListingNamesTheAgentThatKeyedIt(t *testing.T) {
	// arrange
	planned := plan.Plan{Agent: "miso v0.3.1", Stages: []plan.Stage{{Base: "scratch"}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "agent miso v0.3.1\nFROM scratch\n", out.String())
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) {
	return 0, errors.New("the pipe is gone")
}

func TestAListingStopsAtTheFirstWriteError(t *testing.T) {
	// arrange
	planned := plan.Plan{Agent: "miso v0.3.1", Stages: []plan.Stage{{Base: "scratch"}}}

	// act
	err := listing.Write(brokenWriter{}, planned)

	// assert
	assert.EqualError(t, err, "the pipe is gone")
}

func TestAStepWithoutAKeyShowsDashes(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "debian:sid",
		Steps: []plan.Step{{Instruction: imagefile.Run{Line: 2, Command: "debootstrap sid /rootfs"}}},
	}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM debian:sid\n   2  ------------  RUN debootstrap sid /rootfs\n", out.String())
}
