// Package link speaks route netlink with the kernel: it asks for cards,
// addresses and routes in a network namespace and reads the answers. It
// knows nothing of ours and never opens a connection of its own.
package link
