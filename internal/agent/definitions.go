package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/The127/miso/internal/protocol"
)

// writeDefinitions writes the partitions as the files repart reads its
// definitions from, numbered in tens so that they stay in their order.
func writeDefinitions(dir string, partitions []protocol.Partition) error {
	for i, partition := range partitions {
		var text strings.Builder
		text.WriteString("[Partition]\n")

		for _, setting := range partition.Settings {
			text.WriteString(setting.Key + "=" + setting.Value + "\n")
		}

		name := fmt.Sprintf("%02d-%s.conf", (i+1)*10, partition.Name)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text.String()), 0o600); err != nil {
			return err
		}
	}

	return nil
}
