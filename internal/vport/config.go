package vport

import (
	"io"

	"github.com/hashicorp/yamux"
)

// config keeps yamux from logging. Its log would reach miso's standard
// error, and in the VM the console. It keeps no watch on the other side
// either: a VM booting without KVM stays silent for longer than yamux's
// keepalive allows. The builder's patience already bounds the wait, and
// the port closing when QEMU exits ends the session.
func config() *yamux.Config {
	quiet := yamux.DefaultConfig()
	quiet.LogOutput = io.Discard
	quiet.EnableKeepAlive = false

	return quiet
}
