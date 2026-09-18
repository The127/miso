package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/baseimage"
	"github.com/The127/miso/internal/buildcontext"
	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/listing"
	"github.com/The127/miso/internal/plan"
)

var planCommand = &cli.Command{
	Name:      "plan",
	Usage:     "list what a build would do, without building",
	ArgsUsage: "[context]",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: "the build file, Imagefile in the context if not given"},
	},
	Action: listPlan,
}

func listPlan(_ context.Context, command *cli.Command) error {
	dir, file := located(command)

	stages, err := parsed(file)
	if err != nil {
		return err
	}

	files, err := buildcontext.Open(dir)
	if err != nil {
		return err
	}

	defer func() { _ = files.Close() }()

	cache, err := cacheDir()
	if err != nil {
		return err
	}

	bases := baseimage.Open(filepath.Join(cache, "bases"), http.DefaultClient, baseimage.Known)

	planned, err := plan.New(stages, "miso "+built(), files, bases)
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	return listing.Write(command.Root().Writer, planned)
}

// located is the build context and the build file. The context is the
// directory given, or the current one. The file is Imagefile in it, unless
// -f names another.
func located(command *cli.Command) (dir string, file string) {
	dir = command.Args().First()
	if dir == "" {
		dir = "."
	}

	file = command.String("file")
	if file == "" {
		file = filepath.Join(dir, "Imagefile")
	}

	return dir, file
}

// parsed names the file in an error, a line number alone says nothing.
func parsed(file string) ([]imagefile.Stage, error) {
	source, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	stages, err := imagefile.Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}

	return stages, nil
}
