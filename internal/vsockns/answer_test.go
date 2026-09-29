package vsockns_test

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/vsockns"
	"github.com/The127/miso/internal/vsockns/vsocknstest"
)

// kind is what sort of file the helper handed over, or zero when it handed
// over nothing.
func kind(file *os.File, err error) uint32 {
	if err != nil {
		return 0
	}

	defer func() { _ = file.Close() }()

	var stat unix.Stat_t
	if unix.Fstat(int(file.Fd()), &stat) != nil {
		return 0
	}

	return stat.Mode & unix.S_IFMT
}

// asking puts the questions to the namespace, many times, and counts the
// answers that were not what was asked for.
func asking(namespace *vsockns.Namespace, crossed *atomic.Int32) {
	for range 50 {
		if kind(namespace.Socket()) != unix.S_IFSOCK {
			crossed.Add(1)
		}

		if kind(namespace.Device()) != unix.S_IFCHR {
			crossed.Add(1)
		}
	}
}

func TestQuestionsAskedAtOnceGetTheirOwnAnswers(t *testing.T) {
	// arrange
	namespace := vsocknstest.Private(t)
	var crossed atomic.Int32
	var askers sync.WaitGroup

	// act
	for range 8 {
		askers.Go(func() { asking(namespace, &crossed) })
	}

	askers.Wait()

	// assert
	assert.Zero(t, crossed.Load())
}
