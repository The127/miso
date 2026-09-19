package protocol

// Network is what a run needs to reach beyond the builder, as the host
// set up the builder's network.
type Network struct {
	// the MAC the host gave the builder's network card
	Card string
}
