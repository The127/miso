package agent

import (
	"bytes"
	"fmt"
	"os"

	"github.com/The127/miso/internal/vercmp"
)

// cdStub is the first systemd-stub that tells the kernel which partition of
// a CD it was booted from, without which the root on a CD is never found.
const cdStub = "261"

// stubMark is how systemd-stub says its version in its own bytes.
var stubMark = []byte("#### LoaderInfo: systemd-stub ")

// bootsFromCD refuses the UKI stub in the file when it is too old to boot a
// disk from a CD.
func bootsFromCD(stub string) error {
	read, err := os.ReadFile(stub)
	if err != nil {
		return err
	}

	_, after, found := bytes.Cut(read, stubMark)
	if !found {
		return nil
	}

	version, _, _ := bytes.Cut(after, []byte(" ####"))
	if vercmp.Compare(string(version), cdStub) < 0 {
		return fmt.Errorf("an ISO boots from a CD only with systemd-stub %s or newer, the image brings %q", cdStub, version)
	}

	return nil
}
