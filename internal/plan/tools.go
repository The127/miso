package plan

import "github.com/The127/miso/internal/imagefile"

// BuiltinTools names the stage miso adds, with a space that no name in a
// build file can have.
const BuiltinTools = "miso tools"

// withTools puts the stage with the tools first when an output without
// tools of its own needs them.
func withTools(stages []imagefile.Stage) []imagefile.Stage {
	for _, stage := range stages {
		for _, instruction := range stage.Instructions {
			if output, isOutput := instruction.(imagefile.Output); isOutput && makesDisk(output) {
				return append([]imagefile.Stage{builtinStage()}, stages...)
			}
		}
	}

	return stages
}

func makesDisk(output imagefile.Output) bool {
	return output.Kind == "disk" || output.Kind == "iso"
}

// snapshot is the day of the Debian archive the tools come from. The tools
// change when miso moves it, not when Debian does.
const snapshot = "20260928T000000Z"

func builtinStage() imagefile.Stage {
	source := `Types: deb\nURIs: https://snapshot.debian.org/archive/debian/` + snapshot + `/\nSuites: sid\nComponents: main\nCheck-Valid-Until: no\nSigned-By: /usr/share/keyrings/debian-archive-keyring.pgp\n`
	install := imagefile.Run{Command: "rm -f /etc/apt/sources.list /etc/apt/sources.list.d/* && printf '" + source + "' > /etc/apt/sources.list.d/debian.sources && " +
		"apt-get update && apt-get install -y --no-install-recommends systemd-repart systemd-ukify systemd-boot-efi dosfstools mtools e2fsprogs erofs-utils"}

	return imagefile.Stage{Name: BuiltinTools, Base: "debian:sid", Instructions: []imagefile.Instruction{install}}
}
