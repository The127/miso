package version

import "runtime/debug"

// Of is the version a binary was built as.
func Of(info *debug.BuildInfo) string {
	return info.Main.Version
}
