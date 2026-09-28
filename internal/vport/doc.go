// Package vport carries connections between the host and a VM over one
// virtio serial port, for a host without vsock. A port is one lasting
// stream, so the connections are streams multiplexed on it.
package vport
