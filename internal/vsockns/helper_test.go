package vsockns_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/vsockns"
)

func TestAQuestionTheHelperDoesNotKnowIsAnsweredAsFailed(t *testing.T) {
	// arrange
	namespace := opened(t)

	// act
	_, err := vsockns.Ask(namespace, "shutdown")

	// assert
	assert.ErrorContains(t, err, "unknown question shutdown")
}
