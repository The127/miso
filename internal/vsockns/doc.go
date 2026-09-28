// Package vsockns keeps the vsock of miso's VMs to miso. A VM in a vsock
// namespace of local mode cannot be reached from the host's other
// processes, and cannot reach their vsock listeners. Go cannot put a
// running process into a new user namespace, so a helper started from the
// same binary sets the namespace up and does the work inside it.
package vsockns
