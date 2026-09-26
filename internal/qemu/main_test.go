package qemu_test

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/The127/miso/internal/qemu"
)

// TestMain stands in for QEMU when a test starts this binary as one, and
// writes the arguments it was given to the file MISO_FAKE_QEMU names, and a
// line to its console. With MISO_FAKE_QEMU_FAIL it says that on its standard
// error and fails, the way QEMU refuses what it cannot run. With
// MISO_FAKE_QEMU_PID it writes its process ID there. With
// MISO_FAKE_QEMU_HANG it then hangs like a running machine, for a while
// only, so a failed test leaves nothing behind for long.
//
// With MISO_FAKE_MISO it stands in for miso instead, starting a fake QEMU
// and hanging for a while.
func TestMain(m *testing.M) {
	if os.Getenv("MISO_FAKE_MISO") != "" {
		fakeMiso()

		return
	}

	if recorded := os.Getenv("MISO_FAKE_QEMU"); recorded != "" {
		fakeQEMU(recorded)

		return
	}

	// the main thread is the one Go never ends, so no test may run there and
	// miss a thread that ends under it
	runtime.LockOSThread()
	os.Exit(m.Run())
}

func fakeQEMU(recorded string) {
	//nolint:gosec // the test that started this binary names the file
	if err := os.WriteFile(recorded, []byte(strings.Join(os.Args[1:], "\n")), 0o600); err != nil {
		os.Exit(2)
	}

	fmt.Println("fake QEMU console")

	if refusal := os.Getenv("MISO_FAKE_QEMU_FAIL"); refusal != "" {
		fmt.Fprintln(os.Stderr, refusal)
		os.Exit(1)
	}

	if pid := os.Getenv("MISO_FAKE_QEMU_PID"); pid != "" {
		//nolint:gosec // the test that started this binary names the file
		if err := os.WriteFile(pid, []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
			os.Exit(2)
		}
	}

	if os.Getenv("MISO_FAKE_QEMU_HANG") != "" {
		time.Sleep(30 * time.Second)
	}
}

func fakeMiso() {
	// the QEMU it starts is this binary too, and must not be miso again
	if err := os.Unsetenv("MISO_FAKE_MISO"); err != nil {
		os.Exit(2)
	}

	self, err := os.Executable()
	if err != nil {
		os.Exit(2)
	}

	if _, err := (qemu.Driver{Binary: self}).Start(context.Background(), qemu.Machine{}); err != nil {
		os.Exit(2)
	}

	time.Sleep(30 * time.Second)
}
