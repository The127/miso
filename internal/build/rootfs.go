package build

import (
	"slices"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

// rootfs is what a step runs on: the layers under it and the variables set.
type rootfs struct {
	layers []string
	env    []string
}

// base is what the first step of a stage on an image runs on. Scratch has
// nothing in it, not even an empty layer.
func base(stage plan.Stage) rootfs {
	if stage.Base == plan.Scratch {
		return rootfs{}
	}

	return rootfs{layers: []string{stage.BaseKey}}
}

// after is the root file system a step leaves for the steps on top of it.
func (r rootfs) after(step plan.Step) rootfs {
	switch instruction := step.Instruction.(type) {
	case imagefile.Env:
		return rootfs{layers: r.layers, env: withVariable(r.env, instruction)}
	case imagefile.Run, imagefile.Copy:
		return rootfs{layers: slices.Concat(r.layers, []string{step.Key}), env: r.env}
	}

	return r
}
