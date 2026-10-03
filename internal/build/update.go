package build

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"slices"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// ukiFile is the name the agent keeps the UKI of a split disk under.
const ukiFile = "uki.efi"

// splitNameSetting names the file a partition is written to on its own.
const splitNameSetting = "SplitName"

// fetchesOf are the fetches of the files an output keeps: the disk, or the
// files of an update.
func fetchesOf(step plan.Step, output imagefile.Output, partitions []protocol.Partition) []Request {
	if output.Kind == plan.KindUpdate {
		return updateFetches(step, output, partitions)
	}

	line, written := imagefile.Written(output)

	return []Request{{Line: line, Written: written, Message: protocol.Fetch{Key: step.Key}, Output: output.Name, CD: output.Kind == plan.KindISO}}
}

// updateFetches are the fetches of the files an update output ships: one for
// each partition that names its split file and one for the UKI, written
// into the directory the output names as the name and the version say.
func updateFetches(step plan.Step, output imagefile.Output, partitions []protocol.Partition) []Request {
	line, written := imagefile.Written(output)

	var fetches []Request
	for _, partition := range partitions {
		name, named := splitNameOf(partition)
		if !named {
			continue
		}

		file := path.Join(output.Name, name+"_"+output.Options[plan.OptionVersion]+".raw")
		fetches = append(fetches, Request{Line: line, Written: written, Message: protocol.Fetch{Key: step.Key, File: "disk." + name + ".raw"}, Output: file, Listed: true})
	}

	uki := path.Join(output.Name, "uki_"+output.Options[plan.OptionVersion]+".efi")

	return append(fetches, Request{Line: line, Written: written, Message: protocol.Fetch{Key: step.Key, File: ukiFile}, Output: uki, Listed: true})
}

// splitNameOf is the name a partition is split under.
func splitNameOf(partition protocol.Partition) (string, bool) {
	var name string

	for _, setting := range partition.Settings {
		if setting.Key == splitNameSetting {
			name = setting.Value
		}
	}

	return name, name != ""
}

// ErrNoVersion is an update output that says no version, which miso never
// guesses.
var ErrNoVersion = errors.New("an update needs --version")

// ErrBadVersion is a version of an update that systemd-sysupdate cannot read.
var ErrBadVersion = errors.New("not a version of an update")

// ErrBadSplitName is a name a partition is split under that is no plain name.
var ErrBadSplitName = errors.New("not a plain name for a split file")

// ErrDuplicateSplitName is a split name that two partitions share, whose files
// would take one place.
var ErrDuplicateSplitName = errors.New("two partitions have the same SplitName")

// ErrNothingToShip is an update output whose partitions are all kept whole.
var ErrNothingToShip = errors.New("no partition above it has a SplitName, so it ships nothing")

var (
	// the characters of a version of systemd-sysupdate
	versionPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+^~-]*$`)

	// a name that stays one file name in the directory of the update. A
	// lone dash is repart's own word for no split file, so none starts with
	// a dash
	splitNamePattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]*$`)
)

// updateRefusal is why an update output cannot be made of the partitions
// above it, or nil when it can.
func updateRefusal(output imagefile.Output, partitions []protocol.Partition) error {
	version, given := output.Options[plan.OptionVersion]
	if !given {
		return ErrNoVersion
	}

	if !versionPattern.MatchString(version) {
		return fmt.Errorf("--%s=%s: %w", plan.OptionVersion, version, ErrBadVersion)
	}

	var shipped []string

	for _, partition := range partitions {
		name, named := splitNameOf(partition)
		if !named {
			continue
		}

		if !splitNamePattern.MatchString(name) {
			return fmt.Errorf("%s=%s: %w", splitNameSetting, name, ErrBadSplitName)
		}

		if slices.Contains(shipped, name) {
			return fmt.Errorf("%s=%s: %w", splitNameSetting, name, ErrDuplicateSplitName)
		}

		shipped = append(shipped, name)
	}

	if len(shipped) == 0 {
		return ErrNothingToShip
	}

	return nil
}
