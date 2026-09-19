package agent

import "github.com/The127/miso/internal/protocol"

// Serving is what answers the host: the agent started on the cache disk
// with a serial, or one that says why it did not start.
func Serving(serial, dir string) protocol.Runner {
	_, err := Start(serial, dir)

	return Unstarted{Err: err}
}
