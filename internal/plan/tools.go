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
			if output, isOutput := instruction.(imagefile.Output); isOutput && needsBuiltin(output) {
				return append([]imagefile.Stage{builtinStage()}, stages...)
			}
		}
	}

	return stages
}

func needsBuiltin(output imagefile.Output) bool {
	_, hasTools := output.Options[Tools]

	return makesDisk(output) && !hasTools
}

func makesDisk(output imagefile.Output) bool {
	return output.Kind == "disk" || output.Kind == "iso"
}

func builtinStage() imagefile.Stage {
	install := imagefile.Run{Command: "apt-get update && apt-get install -y --no-install-recommends systemd-repart systemd-ukify systemd-boot-efi dosfstools mtools e2fsprogs"}

	return imagefile.Stage{Name: BuiltinTools, Base: "debian:13", Instructions: []imagefile.Instruction{install}}
}
