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

// after is the root file system a step leaves for the steps on top of it.
func (r rootfs) after(step plan.Step) rootfs {
	if variable, isEnv := step.Instruction.(imagefile.Env); isEnv {
		return rootfs{layers: r.layers, env: withVariable(r.env, variable)}
	}

	return rootfs{layers: slices.Concat(r.layers, []string{step.Key}), env: r.env}
}
