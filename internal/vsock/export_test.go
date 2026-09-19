package vsock

func FD(l *Listener) int {
	return l.fd
}
