// Package boot does what the init of a VM does before its work: it moves
// off the initial ramfs, mounts what every program expects and loads the
// kernel modules put next to it. It also powers the VM off.
package boot
