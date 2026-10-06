package builderkernel

import "github.com/The127/miso/internal/download"

// The builder VM boots one pinned Debian cloud kernel of its architecture,
// the same kernel line the VM tests run on. The digest is the guarantee, not
// the archive the bytes come from, so another mirror may serve them later.
var pins = map[string]download.Pin{
	"amd64": {
		URL:    "https://snapshot.debian.org/archive/debian/20260907T023910Z/pool/main/l/linux-signed-amd64/linux-image-6.12.107+deb13-cloud-amd64_6.12.107-1_amd64.deb",
		Digest: "sha256:04434ff520f860423eafe4203d986a35682ce73c2f71baa39e988ca0da93ac91",
	},
	"arm64": {
		URL:    "https://snapshot.debian.org/archive/debian/20260907T023910Z/pool/main/l/linux-signed-arm64/linux-image-6.12.107+deb13-cloud-arm64_6.12.107-1_arm64.deb",
		Digest: "sha256:a1bbfc22454e984feba6353f957152ef416728b627ec0cf94d3ac04b18c14d84",
	},
}
