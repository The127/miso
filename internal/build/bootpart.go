package build

import (
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// bootParts are the parts of the kernel that the outputs of each kind keep.
var bootParts = map[string]string{
	plan.KindKernel: protocol.PartKernel,
	plan.KindInitrd: protocol.PartInitrd,
}
