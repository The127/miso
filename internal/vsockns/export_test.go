package vsockns

import "testing"

// NoChildMode has the helpers a test opens find no way to make their
// namespaces local, as on a host whose kernel has no vsock namespaces.
func NoChildMode(t *testing.T) {
	t.Helper()

	was := childModePath
	childModePath = "/nonexistent/child_ns_mode"
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

// Ask puts any question to the helper.
func Ask(n *Namespace, question string) (string, error) {
	text, _, err := n.ask(question)

	return text, err
}
