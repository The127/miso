package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"

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
		fileFlag(),
		archFlag(),
	},
	Action: listPlan,
}

func listPlan(ctx context.Context, command *cli.Command) error {
	if err := archRefusal(command); err != nil {
		return err
	}

	cache, err := cacheDir()
	if err != nil {
		return err
	}

	blobs, bases, err := baseImages(cache, targetArch(command))
	if err != nil {
		return err
	}

	dir, _ := located(command)
	files, err := buildcontext.Open(dir)
	if err != nil {
		return err
	}

	defer func() { _ = files.Close() }()

	planned, err := planOf(command, files, bases)
	if err != nil {
		return err
	}

	cached, err := cachedKeys(ctx, command, cache, blobs, planned)
	if err != nil {
		return err
	}

	return listing.Write(command.Root().Writer, planned, cached)
}

// planOf is the plan of the build file the command names, on the files of
// its build context and the base images bases knows.
func planOf(command *cli.Command, files *buildcontext.Dir, bases plan.Bases) (plan.Plan, error) {
	_, file := located(command)

	stages, err := parsed(file)
	if err != nil {
		return plan.Plan{}, err
	}

	planned, err := plan.New(stages, agentName(), targetArch(command), files, bases)
	if err != nil {
		return plan.Plan{}, imagefile.InFile(file, err)
	}

	return planned, nil
}

// archFlag names the architecture a command builds for. Each command gets
// its own, since a flag keeps what was parsed into it.
func archFlag() *cli.StringFlag {
	return &cli.StringFlag{Name: "arch", Usage: "the architecture the image is built for", Value: builderArch}
}

// archRefusal is why miso cannot build for the architecture the command
// names, or nil when it can.
func archRefusal(command *cli.Command) error {
	if arch := command.String("arch"); !slices.Contains(arches, arch) {
		return fmt.Errorf("--arch %q is not an architecture miso builds for, only %s", arch, strings.Join(arches, " and "))
	}

	return nil
}

// targetArch is the architecture the command builds for.
func targetArch(command *cli.Command) string {
	return command.String("arch")
}

// fileFlagName is the flag that names the build file.
const fileFlagName = "file"

// fileFlag names the build file of a command that reads one. Each command
// gets its own, since a flag keeps what was parsed into it.
func fileFlag() *cli.StringFlag {
	return &cli.StringFlag{Name: fileFlagName, Aliases: []string{"f"}, Usage: "the build file, Imagefile in the context if not given"}
}

// located is the build context and the build file. The context is the
// directory given, or the current one. The file is Imagefile in it, unless
// -f names another.
func located(command *cli.Command) (dir string, file string) {
	if positional := positionalArgs(command); len(positional) > 0 {
		dir = positional[0]
	}

	if dir == "" {
		dir = "."
	}

	file = command.String(fileFlagName)
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
		return nil, imagefile.InFile(file, err)
	}

	return stages, nil
}
