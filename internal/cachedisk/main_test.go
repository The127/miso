package cachedisk_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/The127/miso/internal/cachedisk"
)

// maker is the name this binary runs under when a test needs Make in a
// process of its own, one it can kill
const maker = "make-cache-disk"

func TestMain(m *testing.M) {
	if os.Args[0] == maker {
		if err := cachedisk.Make(os.Args[1], 64<<20); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	os.Exit(m.Run())
}
