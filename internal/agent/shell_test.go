//go:build vmtest

package agent_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/protocol"
)

func TestAShellKeepsNothingItWrites(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	shell := protocol.Shell{Layers: []string{"base"}}
	in := strings.NewReader("echo hi > /x\nexit\n")

	// act
	code, err := worker.Shell(context.Background(), shell, in, &bytes.Buffer{})

	// assert
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	entries, err := os.ReadDir(layers)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "only the base is left in the layers")
}

func TestAShellSeesTheEnvironmentOfItsStep(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	shell := protocol.Shell{Layers: []string{"base"}, Env: []string{"GREETING=hello"}}
	in := strings.NewReader("echo $GREETING\nexit\n")
	var out bytes.Buffer

	// act
	_, err := worker.Shell(context.Background(), shell, in, &out)

	// assert
	require.NoError(t, err)
	assert.Contains(t, out.String(), "hello\r\n")
}

func TestAShellAnswersItsExitCode(t *testing.T) {
	// arrange
	layers := t.TempDir()
	worker := mountedBase(t, layers)
	shell := protocol.Shell{Layers: []string{"base"}}
	in := strings.NewReader("exit 3\n")

	// act
	code, err := worker.Shell(context.Background(), shell, in, &bytes.Buffer{})

	// assert
	require.NoError(t, err)
	assert.Equal(t, 3, code)
}
