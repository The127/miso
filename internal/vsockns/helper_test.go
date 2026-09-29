package vsockns_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/vsockns"
	"github.com/The127/miso/internal/vsockns/vsocknstest"
)

func TestAQuestionTheHelperDoesNotKnowIsAnsweredAsFailed(t *testing.T) {
	// arrange
	namespace := vsocknstest.Private(t)

	// act
	_, err := vsockns.Ask(namespace, "shutdown")

	// assert
	assert.ErrorContains(t, err, "unknown question shutdown")
}
