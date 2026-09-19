package cachedisk_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/cachedisk"
)

func TestASecondLockWaitsAndSaysSo(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.lock")
	first, err := cachedisk.Lock(path, func() {})
	require.NoError(t, err)

	defer func() { _ = first.Close() }()

	waited := make(chan struct{})

	// act
	go func() {
		second, err := cachedisk.Lock(path, func() { close(waited) })
		if err == nil {
			_ = second.Close()
		}
	}()

	// assert
	select {
	case <-waited:
	case <-time.After(5 * time.Second):
		t.Fatal("the second never said it waits")
	}
}
