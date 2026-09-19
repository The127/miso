//go:build vmtest

package vmtest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTheVMHasANetworkCardWithTheMACTheHostGaveIt(t *testing.T) {
	// arrange
	mac := os.Getenv("MISO_VMTEST_MAC")
	require.NotEmpty(t, mac, "MISO_VMTEST_MAC names no MAC")

	// act
	addresses, err := filepath.Glob("/sys/class/net/*/address")

	// assert
	require.NoError(t, err)
	var found []string
	for _, path := range addresses {
		address, err := os.ReadFile(path)
		require.NoError(t, err)
		found = append(found, strings.TrimSpace(string(address)))
	}

	assert.Contains(t, found, mac)
}
