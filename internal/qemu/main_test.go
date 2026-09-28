package qemu_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/boot"
	"github.com/The127/miso/internal/guestport"
	"github.com/The127/miso/internal/protocol"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vport"
	"github.com/The127/miso/internal/vsock"
)

// TestMain stands in for QEMU when a test starts this binary as one, and
// writes the arguments it was given to the file MISO_FAKE_QEMU names, and a
// line to its console. With MISO_FAKE_QEMU_FAIL it says that on its standard
// error and fails, the way QEMU refuses what it cannot run. With
// MISO_FAKE_QEMU_PID it writes its process ID there. With
// MISO_FAKE_QEMU_HANG it then hangs like a running machine, for a while
// only, so a failed test leaves nothing behind for long. With
// MISO_FAKE_QEMU_ECHO it sends back what reaches it on its first extra file
// until that ends, the way a machine's agent would.
//
// With MISO_FAKE_MISO it stands in for miso instead, starting a fake QEMU
// and hanging for a while.
//
// With MISO_GUEST, which the kernel hands its init from the command line,
// it answers the host over vsock as the init of a VM instead. Being process
// 1 would say that too, but a container's entrypoint is process 1 as well.
func TestMain(m *testing.M) {
	if os.Getenv("MISO_GUEST") != "" {
		guest()

		return
	}

	if os.Getenv("MISO_FAKE_MISO") != "" {
		fakeMiso()

		return
	}

	if recorded := os.Getenv("MISO_FAKE_QEMU"); recorded != "" {
		fakeQEMU(recorded)

		return
	}

	// the main thread is the one Go never ends, so no test may run there and
	// miss a thread that ends under it
	runtime.LockOSThread()
	os.Exit(m.Run())
}

// guestPort is where the guest answers the host.
const guestPort = 1024

// guest readies the VM it is the init of, answers one connection from the
// host with miso, over its virtio port when it has one as the agent does,
// else over vsock, and powers the VM off.
func guest() {
	defer boot.PowerOff()

	if err := boot.Boot(); err != nil {
		fmt.Println(err)

		return
	}

	listener, err := guestListener()
	if err != nil {
		fmt.Println(err)

		return
	}

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err)

		return
	}

	_, _ = fmt.Fprintln(conn, "miso")
	_ = conn.Close()
}

// guestListener mirrors how miso's agent listens, so the guest answers on
// a machine without vsock the way the agent must.
func guestListener() (protocol.Listener, error) {
	device, err := guestport.Find("/sys", "/dev", qemu.AgentPort)
	if errors.Is(err, fs.ErrNotExist) {
		return vsock.Listen(guestPort)
	}

	if err != nil {
		return nil, err
	}

	port, err := os.OpenFile(device, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	return vport.Listen(port)
}

// fakeDriver runs this test binary as its QEMU, which writes the arguments it
// was given to the file whose path comes back.
func fakeDriver(t *testing.T) (qemu.Driver, string) {
	t.Helper()

	recorded := filepath.Join(t.TempDir(), "arguments")
	t.Setenv("MISO_FAKE_QEMU", recorded)
	self, err := os.Executable()
	require.NoError(t, err)

	return qemu.Driver{Binary: self}, recorded
}

func fakeQEMU(recorded string) {
	//nolint:gosec // the test that started this binary names the file
	if err := os.WriteFile(recorded, []byte(strings.Join(os.Args[1:], "\n")), 0o600); err != nil {
		os.Exit(2)
	}

	if temp := os.Getenv("MISO_FAKE_QEMU_TMPDIR"); temp != "" {
		//nolint:gosec // the test that started this binary names the file
		if err := os.WriteFile(temp, []byte(os.Getenv("TMPDIR")), 0o600); err != nil {
			os.Exit(2)
		}
	}

	if held := os.Getenv("MISO_FAKE_QEMU_FDS"); held != "" {
		//nolint:gosec // the test that started this binary names the file
		if err := os.WriteFile(held, []byte(openFiles()), 0o600); err != nil {
			os.Exit(2)
		}
	}

	fmt.Println("fake QEMU console")

	if refusal := os.Getenv("MISO_FAKE_QEMU_FAIL"); refusal != "" {
		_, _ = fmt.Fprintln(os.Stderr, refusal)
		os.Exit(1)
	}

	if pid := os.Getenv("MISO_FAKE_QEMU_PID"); pid != "" {
		//nolint:gosec // the test that started this binary names the file
		if err := os.WriteFile(pid, []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
			os.Exit(2)
		}
	}

	if os.Getenv("MISO_FAKE_QEMU_ECHO") != "" {
		machine := os.NewFile(3, "agent port of QEMU")
		_, _ = io.Copy(machine, machine)
	}

	if os.Getenv("MISO_FAKE_QEMU_HANG") != "" {
		time.Sleep(30 * time.Second)
	}
}

func fakeMiso() {
	// the QEMU it starts is this binary too, and must not be miso again
	if err := os.Unsetenv("MISO_FAKE_MISO"); err != nil {
		os.Exit(2)
	}

	self, err := os.Executable()
	if err != nil {
		os.Exit(2)
	}

	if _, err := (qemu.Driver{Binary: self}).Start(context.Background(), qemu.Machine{}); err != nil {
		os.Exit(2)
	}

	time.Sleep(30 * time.Second)
}

// openFiles are where the open fds of this process point, one a line.
func openFiles() string {
	fds, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return err.Error()
	}

	var targets []string
	for _, fd := range fds {
		target, err := os.Readlink(filepath.Join("/proc/self/fd", fd.Name()))
		if err == nil {
			targets = append(targets, target)
		}
	}

	return strings.Join(targets, "\n")
}
