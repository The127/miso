package plan

import (
	"strings"

	"github.com/The127/miso/internal/imagefile"
)

// BuiltinTools names the stage miso adds, with a space that no name in a
// build file can have.
const BuiltinTools = "miso tools"

// withTools puts the stage with the tools first when an output without
// tools of its own needs them.
func withTools(stages []imagefile.Stage) []imagefile.Stage {
	for _, stage := range stages {
		for _, instruction := range stage.Instructions {
			if output, isOutput := instruction.(imagefile.Output); isOutput && NeedsTools(output.Kind) {
				return append([]imagefile.Stage{builtinStage(output.Line)}, stages...)
			}
		}
	}

	return stages
}

// snapshot is the day of the Debian archive the tools come from. The tools
// change when miso moves it, not when Debian does.
const snapshot = "20260928T000000Z"

// toolPackages are what makes a disk: repart and ukify, the loader and the
// stub they need, and what formats the file systems.
var toolPackages = []string{"systemd-repart", "systemd-ukify", "systemd-boot-efi", "dosfstools", "mtools", "e2fsprogs", "erofs-utils"}

// archive is the apt source of the tools, a line of its file each.
func archive() []string {
	return []string{
		"Types: deb",
		"URIs: https://snapshot.debian.org/archive/debian/" + snapshot + "/",
		"Suites: sid",
		"Components: main",
		"Check-Valid-Until: no",
		"Signed-By: /usr/share/keyrings/debian-archive-keyring.pgp",
	}
}

// installTools is the command that has the tools come from the archive.
func installTools() string {
	source := strings.Join(archive(), `\n`) + `\n`

	return "rm -f /etc/apt/sources.list /etc/apt/sources.list.d/* && " +
		"printf '" + source + "' > /etc/apt/sources.list.d/debian.sources && " +
		"apt-get update && " +
		"apt-get install -y --no-install-recommends " + strings.Join(toolPackages, " ")
}

// builtinStage is on the line of the output that needs it, so that a failure
// in it points somewhere in the build file.
func builtinStage(line int) imagefile.Stage {
	install := imagefile.Run{Line: line, Command: installTools()}

	return imagefile.Stage{Line: line, Name: BuiltinTools, Base: "debian:sid", Instructions: []imagefile.Instruction{install}}
}
