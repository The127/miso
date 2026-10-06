package baseimage

// Known are the images miso can fetch by name for each architecture, cloud
// images as the distributions publish them. A name means whatever is current
// at its address, so a fetch again may bring a different image.
var Known = map[string]map[string]Source{
	"amd64": {
		"debian:sid":    {URL: "https://cloud.debian.org/images/cloud/sid/daily/latest/debian-sid-genericcloud-amd64-daily.qcow2", Format: "qcow2"},
		"debian:13":     {URL: "https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-amd64.qcow2", Format: "qcow2"},
		"debian:trixie": {URL: "https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-amd64.qcow2", Format: "qcow2"},
	},
	"arm64": {
		"debian:sid":    {URL: "https://cloud.debian.org/images/cloud/sid/daily/latest/debian-sid-genericcloud-arm64-daily.qcow2", Format: "qcow2"},
		"debian:13":     {URL: "https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-arm64.qcow2", Format: "qcow2"},
		"debian:trixie": {URL: "https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-arm64.qcow2", Format: "qcow2"},
	},
}
