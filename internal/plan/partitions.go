package plan

import (
	"errors"
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// ErrDuplicatePartition is a partition name that an earlier PARTITION of the
// stage already took, so both would be the same definition file.
var ErrDuplicatePartition = errors.New("partition name taken")

func checkPartitions(stage imagefile.Stage) error {
	taken := map[string]bool{}
	for _, instruction := range stage.Instructions {
		partition, isPartition := instruction.(imagefile.Partition)
		if !isPartition {
			continue
		}

		if !plainName(partition.Name) {
			return at(partition.Line, fmt.Errorf("PARTITION %q: %w", partition.Name, ErrNotAFileName))
		}

		if taken[partition.Name] {
			return at(partition.Line, fmt.Errorf("PARTITION %s: %w", partition.Name, ErrDuplicatePartition))
		}

		taken[partition.Name] = true
	}

	return nil
}
