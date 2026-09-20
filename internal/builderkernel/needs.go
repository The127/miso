package builderkernel

// needs are the modules the builder VM loads, each with what it depends on,
// which the package itself names. A kernel that has one built in is fine,
// the order skips it.
//
// btrfs asks the crypto API for xxhash and blake2b by name when it loads,
// through modprobe, which the builder VM does not have, so the two hashes
// are named here and come first.
var needs = []string{
	// the bus every disk and card of the VM sits on
	"virtio_pci",
	// the cache disk and the base disks of a build
	"virtio_blk",
	// the host's side of a build talks to the agent over vsock
	"vmw_vsock_virtio_transport",
	// the file system the layers live on
	"xxhash_generic",
	"blake2b_generic",
	"btrfs",
	// the root of a run
	"overlay",
	// the card of the VM and the card of a run on it
	"virtio_net",
	"macvlan",
	// the fence a run's network is held by
	"sch_ingress",
	"cls_flower",
	"act_gact",
}
