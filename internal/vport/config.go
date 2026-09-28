package vport

import (
	"io"

	"github.com/hashicorp/yamux"
)

// config keeps yamux from logging. Its log would reach miso's standard
// error, and in the VM the console.
func config() *yamux.Config {
	quiet := yamux.DefaultConfig()
	quiet.LogOutput = io.Discard

	return quiet
}
