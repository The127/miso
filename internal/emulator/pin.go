package emulator

import "github.com/The127/miso/internal/download"

// The builder VM runs the programs of another architecture through one
// pinned Debian qemu-user build of its own architecture. The digest is the
// guarantee, not the archive the bytes come from.
var pins = map[string]download.Pin{
	"amd64": {
		URL:    "https://snapshot.debian.org/archive/debian/20260905T205002Z/pool/main/q/qemu/qemu-user_10.0.13%2Bds-0%2Bdeb13u1_amd64.deb",
		Digest: "sha256:ca6ede739327a20ae5a3498230c8a3a9a072c586e131bc6bcc1201eb1725e336",
	},
	"arm64": {
		URL:    "https://snapshot.debian.org/archive/debian/20260905T205002Z/pool/main/q/qemu/qemu-user_10.0.13%2Bds-0%2Bdeb13u1_arm64.deb",
		Digest: "sha256:c73711af02b97cd5e2667e735d9c06890c57165517110ca4e646c95a7083158d",
	},
}
