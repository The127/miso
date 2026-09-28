package qemu

import (
	"encoding/base64"
	"fmt"
)

// Credential is a systemd credential the machine's PID 1 reads at boot.
type Credential struct {
	Name  string
	Value []byte
}

// credentials are always in base64, so no byte of a value can end the
// SMBIOS string or start the next option.
func credentials(machine Machine) []string {
	var args []string
	for _, credential := range machine.Credentials {
		args = append(args, "-smbios", fmt.Sprintf("type=11,value=io.systemd.credential.binary:%s=%s",
			escaped(credential.Name), base64.StdEncoding.EncodeToString(credential.Value)))
	}

	return args
}
