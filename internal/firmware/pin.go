package firmware

import "github.com/The127/miso/internal/download"

// source is the package of a firmware and the files of its code and vars.
type source struct {
	pin  download.Pin
	code string
	vars string
}

// A check boots through one pinned Debian edk2 build for each architecture,
// whose code and vars match, both without Secure Boot. The digest is the
// guarantee, not the archive the bytes come from.
var sources = map[string]source{
	"amd64": {
		pin: download.Pin{
			URL:    "https://snapshot.debian.org/archive/debian/20260104T203950Z/pool/main/e/edk2/ovmf_2025.02-8%2Bdeb13u1_all.deb",
			Digest: "sha256:78e0d54df11fc77406cb7a0bc9a39e5bca6d1cbe06556b91d9a73491c52decdf",
		},
		// for a 4 MiB flash
		code: "usr/share/OVMF/OVMF_CODE_4M.fd",
		vars: "usr/share/OVMF/OVMF_VARS_4M.fd",
	},
	"arm64": {
		pin: download.Pin{
			URL:    "https://snapshot.debian.org/archive/debian/20260104T203950Z/pool/main/e/edk2/qemu-efi-aarch64_2025.02-8%2Bdeb13u1_all.deb",
			Digest: "sha256:a00b2411a79c8aeafd95a7c868ac3cd1aab592f1af9fff965f02d81a625276ed",
		},
		// AAVMF_CODE.fd is a link to it, which a package reads as empty
		code: "usr/share/AAVMF/AAVMF_CODE.no-secboot.fd",
		vars: "usr/share/AAVMF/AAVMF_VARS.fd",
	},
}
