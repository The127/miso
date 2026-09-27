package builder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/builder"
)

func TestTheRunsLookForTheMACOfTheBuildersCard(t *testing.T) {
	// act
	card, network := builder.Network()

	// assert
	require.NotEmpty(t, card.MAC)
	assert.Equal(t, card.MAC, network.Card)
}
