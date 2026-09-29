package main

import (
	"os"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/builder"
)

const outputFlagName = "output"

func outputFlag() *cli.StringFlag {
	return &cli.StringFlag{Name: outputFlagName, Aliases: []string{"o"}, Usage: "the directory the outputs are written to, none written if not given"}
}

// outputsOf are the outputs in the directory -o names, which is made if it
// is missing. A disk with checks is booted through check before it gets its
// name. Without -o only the disks with checks are fetched, into scratch, so
// the checks still run.
func outputsOf(command *cli.Command, check builder.Check, scratch string) (builder.Outputs, error) {
	dir := command.String(outputFlagName)
	if dir == "" {
		return builder.OnlyChecked(builder.OutputsIn(scratch, check)), nil
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}

	return builder.OutputsIn(dir, check), nil
}
