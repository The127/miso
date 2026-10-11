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

// emulators are the digests of the programs in each package, by the
// builder's architecture and then the architecture a program runs the
// programs of. A build keeps the one it takes out under its digest, and
// later builds read it without opening the package again.
var emulators = map[string]map[string]string{
	"amd64": {
		"amd64": "sha256:36d35ffe5363bd49bfa9c2b6a048e9fb2e798fea38285a5a45f52d34f9869089",
		"arm64": "sha256:73cc2584f119f9e74c85c0a8549dbb6fd104ae3897a06a7500797f9e5b4d0943",
	},
	"arm64": {
		"amd64": "sha256:9af86352fd463ad9ec25bdd6a4b891ac6b6a7f3d46a7db3da6f23e2c09094a59",
		"arm64": "sha256:088500814a5a8e0e0ee3cad0f6c600b5dfa261f77cc0e6460f81b0b616e09509",
	},
}
