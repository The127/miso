package cachedisk_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestASecondLockGetsTheCacheOnceTheFirstLetsGo(t *testing.T) {
	// arrange
	path := filepath.Join(t.TempDir(), "layers.lock")
	first, err := cachedisk.Lock(path, func() {})
	require.NoError(t, err)
	waited := make(chan struct{})
	got := make(chan error, 1)

	go func() {
		second, err := cachedisk.Lock(path, func() { close(waited) })
		if err == nil {
			err = second.Close()
		}

		got <- err
	}()

	<-waited

	// act
	require.NoError(t, first.Close())

	// assert
	select {
	case err := <-got:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("the second never got the cache")
	}
}
