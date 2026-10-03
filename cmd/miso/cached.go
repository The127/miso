package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// cachedKeys are the keys of the steps of a plan whose layers are on the
// cache disk. A cache with no disk yet holds none, and a plan with no keys,
// whose bases are not fetched yet, asks for none. The builder does not boot
// for either.
func cachedKeys(ctx context.Context, command *cli.Command, cache string, blobs *download.Store, planned plan.Plan) (map[string]bool, error) {
	var keys []string

	for _, stage := range planned.Stages {
		for _, step := range stage.Steps {
			if step.Key != "" {
				keys = append(keys, step.Key)
			}
		}
	}

	dir := builderDir(cache)
	if _, err := os.Stat(dir.Disk()); len(keys) == 0 || errors.Is(err, fs.ErrNotExist) {
		return map[string]bool{}, nil
	}

	held, _, err := lockedDisk(dir, command.Root().ErrWriter)
	if err != nil {
		return nil, err
	}

	defer func() { _ = held.Close() }()

	var said bytes.Buffer
	if err := askBuilder(ctx, command, dir, blobs, protocol.Cached{Keys: keys}, &said); err != nil {
		return nil, err
	}

	cached := map[string]bool{}
	for key := range strings.SplitSeq(said.String(), "\n") {
		if key != "" {
			cached[key] = true
		}
	}

	return cached, nil
}
