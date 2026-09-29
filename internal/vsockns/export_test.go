package vsockns

import (
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// NoChildMode has the helpers a test opens find no way to make their
// namespaces local, as on a host whose kernel has no vsock namespaces.
func NoChildMode(t *testing.T) {
	t.Helper()

	was := childModePath
	childModePath = "/nonexistent/child_ns_mode"
	t.Cleanup(func() { childModePath = was })
}

// UnheededChildMode has the helpers a test opens write their mode into a
// file that changes nothing, as on a kernel that takes the write but not the
// mode.
func UnheededChildMode(t *testing.T) {
	t.Helper()

	was := childModePath
	childModePath = filepath.Join(t.TempDir(), "child_ns_mode")
	t.Cleanup(func() { childModePath = was })
}

// NoDevice has the helpers a test opens find no vsock device, as on a host
// without the vhost_vsock module.
func NoDevice(t *testing.T) {
	t.Helper()

	was := devicePath
	devicePath = "/nonexistent/vhost-vsock"
	t.Cleanup(func() { devicePath = was })
}

// IsHelper is whether a program with the arguments and the file as its
// channel is the helper.
var IsHelper = isHelper

// HangingChildMode has the helpers a test opens hang before their first
// answer, on a FIFO nobody reads, and names the FIFO.
func HangingChildMode(t *testing.T) string {
	t.Helper()

	fifo := filepath.Join(t.TempDir(), "child_ns_mode")
	if err := unix.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}

	was := childModePath
	childModePath = fifo
	t.Cleanup(func() {
		childModePath = was
		// a helper still stuck would outlive the test
		if reader, err := unix.Open(fifo, unix.O_RDONLY|unix.O_NONBLOCK, 0); err == nil {
			_ = unix.Close(reader)
		}
	})

	return fifo
}

// Impatient has Open give up on a helper after a moment.
func Impatient(t *testing.T) {
	t.Helper()

	was := openPatience
	openPatience = 100 * time.Millisecond
	t.Cleanup(func() { openPatience = was })
}

// Conn is the number of miso's end of the channel to the helper.
func Conn(n *Namespace) int {
	return n.conn
}

// Ask puts any question to the helper.
func Ask(n *Namespace, question string) (string, error) {
	text, _, err := n.ask(question)

	return text, err
}
