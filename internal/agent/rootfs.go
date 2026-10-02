package agent

import (
	"context"
	"io"
	"path/filepath"

	"github.com/The127/miso/internal/protocol"
)

// mkfs makes an ext4 file system of the image. mkfs cannot grow its file
// and picks its inodes from the size, so both are counted from the image.
// Every entry takes a block of its own, which the sizes of the files do not
// show.
const mkfs = `set -e
entries=$(du -s --inodes /run/miso/image | cut -f1)
size=$(du -sk --apparent-size /run/miso/image | cut -f1)
size=$((size + entries * 4))
truncate -s "$((size + size / 4 + 65536))K" ` + outputPath + `
mkfs.ext4 -q -F -N "$((entries + entries / 4 + 1024))" -d /run/miso/image ` + outputPath

// Rootfs makes an ext4 file system of the image's layers with the tools of
// another stage. A key whose layer is there already has its file system.
func (a *Agent) Rootfs(ctx context.Context, request protocol.Rootfs, out io.Writer) error {
	return a.madeByTools(request.Key, request.Layers, request.Tools, "file system", func(dir toolsDir) (toolsRun, error) {
		return func(root string) (int, error) {
			unbindOutput, err := bind(dir.output, filepath.Join(root, "run", "miso", "out"))
			if err != nil {
				return 0, err
			}

			defer unbindOutput()

			return overImage(ctx, root, dir.booting, dir.image.below, protocol.Run{Command: mkfs}, out)
		}, nil
	})
}
