package vercmp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/vercmp"
)

func TestTheBiggerNumberComparesHigher(t *testing.T) {
	// act
	compared := vercmp.Compare("9", "10")

	// assert
	assert.Negative(t, compared)
}
