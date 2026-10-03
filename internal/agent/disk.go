package agent

import (
	"context"
	"io"
	"path/filepath"

	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/sandbox"
)

// systemdBoot is where an image brings systemd-boot.
const systemdBoot = "usr/lib/systemd/boot/efi/systemd-bootx64.efi"

// stub is where an image brings the stub a UKI is built on.
const stub = "usr/lib/systemd/boot/efi/linuxx64.efi.stub"

// repart makes the disk of the image from the definitions the image ships,
// with a boot catalog for optical drives when asked. Without --dry-run=no
// it writes nothing. OVMF cannot read an ESP of 4096 byte sectors, which
// repart may pick for a file.
const repart = `set --
if [ -n "$MISO_EL_TORITO" ]; then
	set -- --el-torito=yes
fi
if [ -n "$MISO_SPLIT" ]; then
	set -- "$@" --split=yes
fi
if [ -d /run/miso/definitions ]; then
	set -- --definitions=/run/miso/definitions "$@"
fi
systemd-repart \
	--dry-run=no \
	--root=/run/miso/image \
	--offline=yes \
	--sector-size=512 \
	--empty=create \
	--size=auto \
	"$@" \
	` + outputPath

// ukify builds the UKI of the image's kernel into the ESP from copies of
// the image's parts, with its kernel command line when it has one.
const ukify = `set --
if [ -e /run/miso/parts/cmdline ]; then
	set -- --cmdline=@/run/miso/parts/cmdline
fi
ukify build \
	--linux=/run/miso/parts/linux \
	--initrd=/run/miso/parts/initrd \
	--stub=/run/miso/parts/stub \
	--os-release=@/run/miso/parts/os-release \
	--uname="$MISO_VERSION" \
	--output="/run/miso/esp/EFI/Linux/$MISO_VERSION.efi" \
	"$@"`

// Disk makes a bootable disk image of the image's layers with the tools of
// another stage. A key whose layer is there already has its disk.
func (a *Agent) Disk(ctx context.Context, request protocol.Disk, out io.Writer) error {
	return a.madeByTools(request.Key, request.Layers, request.Tools, "disk", func(dir toolsDir) (toolsRun, error) {
		// links in the image mean places in the image, never in the builder VM
		imageFS := place.Open(dir.image.dir).FS()

		boot, err := prepareBoot(imageFS, dir.image.dir, dir.scratch, request)
		if err != nil {
			return nil, err
		}

		// without partitions the image's own definitions are the layout
		var definitions string
		if len(request.Partitions) > 0 {
			definitions, err = makeDefinitions(dir.scratch, request.Partitions)
			if err != nil {
				return nil, err
			}
		}

		in := toolsInput{
			output:      dir.output,
			definitions: definitions,
			boot:        boot,
			booting:     dir.booting,
			// the ESP is whole once the UKI is in it, and no lower layer may
			// change under a mounted overlay
			below:    append([]string{boot.esp}, dir.image.below...),
			elTorito: request.ElTorito,
			split:    request.Split,
		}

		return func(root string) (int, error) {
			code, err := runTools(ctx, root, in, out)
			if err != nil || code != 0 || !request.Split {
				return code, err
			}

			return 0, keepUKI(boot, dir.output)
		}, nil
	})
}

// toolsInput is what the tools in their root are shown to make a disk.
type toolsInput struct {
	output      string
	definitions string
	boot        boot
	booting     string
	below       []string
	elTorito    bool
	split       bool
}

// runTools has the tools in root build the UKI and make the disk.
func runTools(ctx context.Context, root string, in toolsInput, out io.Writer) (int, error) {
	unbind, err := bindForTools(root, in.output, in.definitions)
	if err != nil {
		return 0, err
	}

	defer unbind()

	code, err := buildUKI(ctx, root, in.boot.parts, in.boot.esp, in.boot.kernel.Version, out)
	if err != nil || code != 0 {
		return code, err
	}

	run := protocol.Run{Command: repart}
	if in.elTorito {
		run.Env = append(run.Env, "MISO_EL_TORITO=1")
	}

	if in.split {
		run.Env = append(run.Env, "MISO_SPLIT=1")
	}

	return overImage(ctx, root, in.booting, in.below, run, out)
}

// buildUKI has the tools in root build the UKI of the kernel of a version
// into the ESP, from copies of the image's boot parts.
func buildUKI(ctx context.Context, root, parts, esp, version string, out io.Writer) (int, error) {
	unbindParts, err := bind(parts, filepath.Join(root, "run", "miso", "parts"))
	if err != nil {
		return 0, err
	}

	defer unbindParts()

	unbindESP, err := bind(filepath.Join(esp, "efi"), filepath.Join(root, "run", "miso", "esp"))
	if err != nil {
		return 0, err
	}

	defer unbindESP()

	// a version from the image reaches the shell only as a value, never as
	// its words
	return sandbox.Run(ctx, root, protocol.Run{Command: ukify, Env: []string{"MISO_VERSION=" + version}}, out)
}
