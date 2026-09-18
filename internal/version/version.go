package version

import "runtime/debug"

// Of is the version a binary was built as. A build from a checkout has no
// version, only the commit it was built from.
func Of(info *debug.BuildInfo) string {
	if info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	revision := setting(info, "vcs.revision")
	if revision == "" {
		return "unknown"
	}

	return revision[:12]
}

func setting(info *debug.BuildInfo, key string) string {
	for _, found := range info.Settings {
		if found.Key == key {
			return found.Value
		}
	}

	return ""
}
