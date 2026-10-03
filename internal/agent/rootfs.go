package agent

import (
	"context"
	"fmt"
	"io"
	"strings"

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

// requiresOsRelease refuses an image portablectl would refuse to attach. A
// link in the image may point at a place of the tools, so a link counts as
// there, even a dangling one.
const requiresOsRelease = `for file in usr/lib/os-release etc/os-release; do
	if [ -e "/run/miso/image/$file" ] || [ -L "/run/miso/image/$file" ]; then
		found=1
	fi
done
if [ -z "$found" ]; then
	echo "miso: a portable image needs /usr/lib/os-release or /etc/os-release" >&2
	exit 1
fi
`

// requiresExtensionRelease refuses an extension systemd would not merge,
// which is one without the release file of its name in the directory of its
// kind. A link counts as there for the reason above.
func requiresExtensionRelease(kind, dir string) string {
	return fmt.Sprintf(`file="/run/miso/image/%[2]s/extension-release.d/extension-release.$MISO_IMAGE"
if [ ! -e "$file" ] && [ ! -L "$file" ]; then
	echo "miso: a %[1]s needs /%[2]s/extension-release.d/extension-release.$MISO_IMAGE" >&2
	exit 1
fi
`, kind, dir)
}

// makers are the scripts that make the file systems a rootfs can be.
var makers = map[string]string{
	protocol.FormatExt4:  mkfsExt4,
	protocol.FormatErofs: mkfsErofs,
}

// wraps are the checks of the images a file system can be wrapped in, which
// run before repart.
var wraps = map[string]string{
	protocol.WrapPortable: requiresOsRelease,
	protocol.WrapSysext:   requiresExtensionRelease("sysext", "usr/lib"),
	protocol.WrapConfext:  requiresExtensionRelease("confext", "etc"),
}

// Rootfs makes a file system of the image's layers with the tools of
// another stage. A key whose layer is there already has its file system.
func (a *Agent) Rootfs(ctx context.Context, request protocol.Rootfs, out io.Writer) error {
	maker, known := makers[request.Format]
	if !known {
		return fmt.Errorf("no file system %q can be made", request.Format)
	}

	if request.Wrap != "" {
		return a.wrapped(ctx, request, out)
	}

	return a.madeByTools(request.Key, request.Layers, request.Tools, "file system", func(dir toolsDir) (toolsRun, error) {
		return func(root string) (int, error) {
			unbind, err := bindForTools(root, dir.output, "")
			if err != nil {
				return 0, err
			}

			defer unbind()

			return overImage(ctx, root, dir.booting, dir.image.below, protocol.Run{Command: maker}, out)
		}, nil
	})
}

// wrapped makes a file system of the image's layers, wrapped in a disk of
// its own by repart, after the checks of the kind of image.
func (a *Agent) wrapped(ctx context.Context, request protocol.Rootfs, out io.Writer) error {
	requirements, known := wraps[request.Wrap]
	if !known {
		return fmt.Errorf("no image of kind %q can be made", request.Wrap)
	}

	return a.madeByTools(request.Key, request.Layers, request.Tools, "file system", func(dir toolsDir) (toolsRun, error) {
		definitions, err := makeDefinitions(dir.scratch, []protocol.Partition{wrappedPartition(request.Format)})
		if err != nil {
			return nil, err
		}

		run := protocol.Run{Command: requirements + repart, Env: []string{"MISO_IMAGE=" + strings.TrimSuffix(request.Name, ".raw")}}

		return func(root string) (int, error) {
			unbind, err := bindForTools(root, dir.output, definitions)
			if err != nil {
				return 0, err
			}

			defer unbind()

			return overImage(ctx, root, dir.booting, dir.image.below, run, out)
		}, nil
	})
}

// wrappedPartition is the one partition of a wrapped image, which holds the
// whole image. Without Minimize repart sizes an ext4 too small, and best
// is refused for ext4. Type=root names the root of x86-64 only, which is
// the one architecture miso builds for.
func wrappedPartition(format string) protocol.Partition {
	return protocol.Partition{Name: "root", Settings: []protocol.Setting{
		{Key: "Type", Value: "root"},
		{Key: "Format", Value: format},
		{Key: "CopyFiles", Value: "/"},
		{Key: "Minimize", Value: "guess"},
	}}
}
