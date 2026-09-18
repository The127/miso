package baseimage

// Known are the images miso can fetch by name, amd64 cloud images as the
// distributions publish them. A name means whatever is current at its
// address, so a fetch again may bring a different image.
var Known = map[string]string{
	"debian:sid":    "https://cloud.debian.org/images/cloud/sid/daily/latest/debian-sid-genericcloud-amd64-daily.qcow2",
	"debian:13":     "https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-amd64.qcow2",
	"debian:trixie": "https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-amd64.qcow2",
}
