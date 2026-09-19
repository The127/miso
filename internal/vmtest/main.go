package vmtest

import (
	"fmt"
	"os"
	"testing"

	"github.com/The127/miso/internal/boot"
)

// Main runs the tests of a package as the init of a VM, then powers the VM
// off. The last line it prints tells the host how they went.
func Main(m *testing.M) {
	if os.Getpid() != 1 {
		fmt.Fprintln(os.Stderr, "these tests run as the init of a VM, see just test-vm")
		os.Exit(1)
	}

	code := 1
	if err := boot.Boot(); err != nil {
		fmt.Println(err)
	} else {
		code = m.Run()
	}

	fmt.Printf("miso-vmtest: exit %d\n", code)
	boot.PowerOff()
}
