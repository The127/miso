package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/plan"
)

func TestAPlanKeepsTheBaseDigestItWasKeyedWith(t *testing.T) {
	// arrange
	stages := parse(t, "FROM debian:sid\nRUN true\n")

	// act
	planned, err := plan.New(stages, anyAgent, noFiles, images{"debian:sid": "sha256:old"})

	// assert
	require.NoError(t, err)
	require.Len(t, planned.Stages, 1)
	assert.Equal(t, "sha256:old", planned.Stages[0].BaseDigest)
}
