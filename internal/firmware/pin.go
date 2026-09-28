package firmware

import "github.com/The127/miso/internal/download"

// A check boots through one pinned Debian OVMF build, whose code and vars
// match. The digest is the guarantee, not the archive the bytes come from.
var pin = download.Pin{
	URL:    "https://snapshot.debian.org/archive/debian/20260104T203950Z/pool/main/e/edk2/ovmf_2025.02-8%2Bdeb13u1_all.deb",
	Digest: "sha256:78e0d54df11fc77406cb7a0bc9a39e5bca6d1cbe06556b91d9a73491c52decdf",
}
