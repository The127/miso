package agent

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/The127/miso/internal/protocol"
)

// mkfsExt4 makes an ext4 file system of the image. mkfs cannot grow its file
// and picks its inodes from the size, so both are counted from the image.
// Every entry takes a block of its own, which the sizes of the files do not
// show.
const mkfsExt4 = `set -e
entries=$(du -s --inodes /run/miso/image | cut -f1)
size=$(du -sk --apparent-size /run/miso/image | cut -f1)
size=$((size + entries * 4))
truncate -s "$((size + size / 4 + 65536))K" ` + outputPath + `
mkfs.ext4 -q -F -N "$((entries + entries / 4 + 1024))" -d /run/miso/image ` + outputPath

// mkfsErofs makes an erofs file system of the image, which sizes its file
// itself.
const mkfsErofs = `mkfs.erofs ` + outputPath + ` /run/miso/image`

// makers are the scripts that make the file systems a rootfs can be.
var makers = map[string]string{
	protocol.FormatExt4:  mkfsExt4,
	protocol.FormatErofs: mkfsErofs,
}

// Rootfs makes a file system of the image's layers with the tools of
// another stage. A key whose layer is there already has its file system.
func (a *Agent) Rootfs(ctx context.Context, request protocol.Rootfs, out io.Writer) error {
	maker, known := makers[request.Format]
	if !known {
		return fmt.Errorf("no file system %q can be made", request.Format)
	}

	return a.madeByTools(request.Key, request.Layers, request.Tools, "file system", func(dir toolsDir) (toolsRun, error) {
		command := maker

		var definitions string

		if request.Portable {
			var err error

			definitions, err = makeDefinitions(dir.scratch, []protocol.Partition{portablePartition(request.Format)})
			if err != nil {
				return nil, err
			}

			command = repart
		}

		return func(root string) (int, error) {
			unbindOutput, err := bindOutput(root, dir.output)
			if err != nil {
				return 0, err
			}

			defer unbindOutput()

			if request.Portable {
				unbindDefinitions, err := bind(definitions, filepath.Join(root, "run", "miso", "definitions"))
				if err != nil {
					return 0, err
				}

				defer unbindDefinitions()
			}

			return overImage(ctx, root, dir.booting, dir.image.below, protocol.Run{Command: command}, out)
		}, nil
	})
}

// portablePartition is the one partition of a portable image, which holds
// the whole image. Without Minimize repart sizes an ext4 too small, and best
// is refused for ext4. Type=root names the root of x86-64 only, which is
// the one architecture miso builds for.
func portablePartition(format string) protocol.Partition {
	return protocol.Partition{Name: "root", Settings: []protocol.Setting{
		{Key: "Type", Value: "root"},
		{Key: "Format", Value: format},
		{Key: "CopyFiles", Value: "/"},
		{Key: "Minimize", Value: "guess"},
	}}
}
