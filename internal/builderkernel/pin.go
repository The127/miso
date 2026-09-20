package builderkernel

// The builder VM boots one pinned Debian cloud kernel, the same kernel line
// the VM tests run on. The digest is the guarantee, not the archive the
// bytes come from, so another mirror may serve them later.
const (
	pinURL    = "https://snapshot.debian.org/archive/debian/20260907T023910Z/pool/main/l/linux-signed-amd64/linux-image-6.12.107+deb13-cloud-amd64_6.12.107-1_amd64.deb"
	pinDigest = "sha256:04434ff520f860423eafe4203d986a35682ce73c2f71baa39e988ca0da93ac91"
)
