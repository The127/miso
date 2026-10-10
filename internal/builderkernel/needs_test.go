package builderkernel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/builderkernel"
)

func TestTheBuilderVMLoadsTheDriverOfItsAgentsVirtioPort(t *testing.T) {
	// arrange
	driver := "virtio_console"

	// act
	loads := builderkernel.Needs

	// assert
	assert.Contains(t, loads, driver)
}

func TestTheBuilderVMLoadsWhatRunsProgramsOfAnotherArchitecture(t *testing.T) {
	// arrange
	module := "binfmt_misc"

	// act
	loads := builderkernel.Needs

	// assert
	assert.Contains(t, loads, module)
}
