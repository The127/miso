package listing_test

import (
	"bytes"
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

func TestAnEnvStepReadsAsWritten(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Env{Line: 3, Key: "LANG", Value: "C.UTF-8"}, Key: "a1b2c3d4e5f6"}},
	}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM scratch\n   3  a1b2c3d4e5f6  ENV LANG=C.UTF-8\n", out.String())
}

func TestACopyStepReadsAsWritten(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Copy{Line: 2, Sources: []string{"etc/motd", "etc/issue"}, Destination: "/etc/"}, Key: "9a8b7c6d5e4f"}},
	}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM scratch\n   2  9a8b7c6d5e4f  COPY etc/motd etc/issue /etc/\n", out.String())
}

func TestACopyFromAStageShowsTheStage(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Copy{Line: 5, From: "build", Sources: []string{"/out/app"}, Destination: "/usr/bin/"}, Key: "9a8b7c6d5e4f"}},
	}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM scratch\n   5  9a8b7c6d5e4f  COPY --from=build /out/app /usr/bin/\n", out.String())
}

func TestAnOutputStepShowsItsOptionsInOneOrder(t *testing.T) {
	// arrange
	planned := plan.Plan{Stages: []plan.Stage{{
		Base:  "scratch",
		Steps: []plan.Step{{Instruction: imagefile.Output{Line: 7, Kind: "disk", Name: "os.img", Options: map[string]string{"verity": "", "size": "4G"}}, Key: "9a8b7c6d5e4f"}},
	}}}
	var out bytes.Buffer

	// act
	err := listing.Write(&out, planned)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "FROM scratch\n   7  9a8b7c6d5e4f  OUTPUT disk os.img --size=4G --verity\n", out.String())
}
