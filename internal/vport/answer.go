package vport

import (
	"fmt"
	"time"

	"github.com/hashicorp/yamux"
)

// pingPatience is how long Dial waits for the VM's side of the port to
// answer. A
// yamux stream opens without the other side, but the host must learn that
// no agent listens yet and dial again, as it does over vsock.
const pingPatience = time.Second

// answered fails unless the VM's side of the session answers a ping. A
// ping holds none of the session's streams, so dialing before the agent
// runs leaves nothing behind for it.
func answered(session *yamux.Session) error {
	pinged := make(chan error, 1)
	go func() {
		_, err := session.Ping()
		pinged <- err
	}()

	select {
	case err := <-pinged:
		return err
	case <-time.After(pingPatience):
		return fmt.Errorf("no agent answered on the port within %s", pingPatience)
	}
}
