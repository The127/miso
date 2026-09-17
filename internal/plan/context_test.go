package plan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/plan"
)

// files is a context in memory, a file's content stands in for its digest.
type files map[string]string

func (f files) Digest(path string) string {
	return f[path]
}

var noFiles = files{}

func TestAChangedContextFileChangesTheKeyOfItsCopy(t *testing.T) {
	// arrange
	stages := parse(t, "FROM scratch\nCOPY motd /etc/\n")

	// act
	helloKeys := plan.Keys(stages, anyAgent, files{"motd": "hello"})
	goodbyeKeys := plan.Keys(stages, anyAgent, files{"motd": "goodbye"})

	// assert
	assert.NotEqual(t, lastKey(t, helloKeys), lastKey(t, goodbyeKeys))
}
