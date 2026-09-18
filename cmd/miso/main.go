package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/buildcontext"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/listing"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/version"
)

func main() {
	miso := &cli.Command{
		Name:  "miso",
		Usage: "build systemd-based operating systems from a build file",
		Commands: []*cli.Command{
			{
				Name:      "plan",
				Usage:     "list what a build would do, without building",
				ArgsUsage: "[context]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: "the build file, Imagefile in the context if not given"},
				},
				Action: listPlan,
			},
			{
				Name:   "version",
				Usage:  "say which miso this is",
				Action: printVersion,
			},
		},
	}

	if err := miso.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "miso:", err)
		os.Exit(1)
	}
}

func listPlan(_ context.Context, command *cli.Command) error {
	dir := command.Args().First()
	if dir == "" {
		dir = "."
	}

	file := command.String("file")
	if file == "" {
		file = filepath.Join(dir, "Imagefile")
	}

	source, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	stages, err := imagefile.Parse(string(source))
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	files, err := buildcontext.Open(dir)
	if err != nil {
		return err
	}

	defer func() { _ = files.Close() }()

	planned, err := plan.New(stages, agent(), files, noBases{})
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	return listing.Write(command.Root().Writer, planned)
}

// noBases stands in until miso can fetch base images.
type noBases struct{}

func (noBases) Digest(string) (string, error) {
	return "", errors.New("no base images yet, a FROM takes scratch or an earlier stage")
}

// agent is what a key says about the miso that made it.
func agent() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "miso unknown"
	}

	return "miso " + version.Of(info)
}

func printVersion(_ context.Context, command *cli.Command) error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New("this binary carries no build information")
	}

	_, err := fmt.Fprintln(command.Root().Writer, version.Of(info))

	return err
}
