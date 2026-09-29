package vsockns

import (
	"path/filepath"
	"testing"
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

// Conn is the number of miso's end of the channel to the helper.
func Conn(n *Namespace) int {
	return n.conn
}

// Ask puts any question to the helper.
func Ask(n *Namespace, question string) (string, error) {
	text, _, err := n.ask(question)

	return text, err
}
