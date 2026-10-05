package qemu

import (
	"encoding/base64"
	"fmt"
	"regexp"
)

// Credential is a systemd credential the machine's PID 1 reads at boot.
type Credential struct {
	Name  string
	Value []byte
}

// credentials are always in base64, so no byte of a value can end the
// SMBIOS string or start the next option.
func credentials(machine Machine) []string {
	if !smbios(machine) {
		return nil
	}

	var args []string
	for _, credential := range machine.Credentials {
		args = append(args, "-smbios", fmt.Sprintf("type=11,value=io.systemd.credential.binary:%s=%s",
			escaped(credential.Name), base64.StdEncoding.EncodeToString(credential.Value)))
	}

	return args
}

// commandLineCredentials are the credentials as words of the kernel command
// line, which systemd reads from there. systemd does not see them in SMBIOS
// on a microvm, and a name of a credential is too long for the names of
// fw_cfg. A name is only ever the letters of a unit name, so no word of the
// line can be split or cut by it.
func commandLineCredentials(machine Machine) ([]string, error) {
	words := make([]string, 0, len(machine.Credentials))
	for _, credential := range machine.Credentials {
		if !credentialName.MatchString(credential.Name) {
			return nil, fmt.Errorf("the credential %q has characters that a word of the kernel command line cannot have", credential.Name)
		}

		words = append(words, fmt.Sprintf("systemd.set_credential_binary=%s:%s",
			credential.Name, base64.StdEncoding.EncodeToString(credential.Value)))
	}

	return words, nil
}

// credentialName is what a name of a credential on the command line is made of.
var credentialName = regexp.MustCompile(`^[A-Za-z0-9._@-]+$`)

// smbios says whether the machine has a firmware that hands systemd an SMBIOS
// table, its board's own or the one it boots.
func smbios(machine Machine) bool {
	_, bootsFirmware := machine.Boot.(Firmware)

	return boardOf(machine).firmware || bootsFirmware
}
