package main_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	main "github.com/The127/miso/cmd/miso"
	"github.com/The127/miso/internal/imagefile"
)

// hinted is the error of a failed step on line 2 with its hint, as a build
// run with the arguments gives it.
func hinted(t *testing.T, args ...string) error {
	t.Helper()

	var got error
	build := &cli.Command{
		Name:  "build",
		Flags: []cli.Flag{main.FileFlag(), main.ArchFlag()},
		Action: func(_ context.Context, command *cli.Command) error {
			got = main.WithShellHint(command, nil, &imagefile.Error{Line: 2, Err: errors.New("RUN false: exit code 1")})

			return nil
		},
	}
	require.NoError(t, build.Run(context.Background(), append([]string{"build"}, args...)))

	return got
}

func TestTheShellAFailingStepOfABuildForTheOtherArchitectureSaysToOpenIsForThatArchitecture(t *testing.T) {
	// arrange
	other := "arm64"
	if main.BuilderArch == "arm64" {
		other = "amd64"
	}

	// act
	err := hinted(t, "--arch", other)

	// assert
	assert.ErrorContains(t, err, "miso shell --before 2 --arch "+other)
}
