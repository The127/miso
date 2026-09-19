package protocol

// Network is what a run needs to reach beyond the builder, as the host
// set up the builder's network.
type Network struct {
	// the MAC the host gave the builder's network card
	Card string

	IPv4 Family
	IPv6 Family
}

// Family is the part of a run's network in one address family.
type Family struct {
	// the run's own, with its prefix length, 10.0.2.15/24
	Address string
	Gateway string
}
