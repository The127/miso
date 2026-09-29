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
// is missing, or none without -o.
func outputsOf(command *cli.Command) (builder.Outputs, error) {
	dir := command.String(outputFlagName)
	if dir == "" {
		return nil, nil
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}

	return builder.OutputsIn(dir, nil), nil
}
