# Checks

A `CHECK` boots an output and runs a command in the booted system. If the
command fails, the build fails and the output is not written. This is how
a build proves that the image it made starts.

```
OUTPUT disk os.raw

CHECK command -v htop
```

## Which outputs can be checked

A `CHECK` belongs to the closest `OUTPUT` above it in the same stage. It can
check:

- a `disk`
- an `iso`
- a `rootfs`

The other kinds are never booted. A `CHECK` after one of them is an error.
A `CHECK` with no `OUTPUT` above it is an error too.

The image has to run systemd. miso learns that the system has booted from
systemd, and it runs the commands through a systemd service.

### Checking a rootfs

A rootfs has no kernel, so miso boots the kernel and the initrd of the stage
with the rootfs as a virtual disk. The stage needs, above the `OUTPUT rootfs`:

- an `OUTPUT kernel`
- an `OUTPUT initrd`
- a `CMDLINE` with a `root=` word

```
CMDLINE root=/dev/vda rootfstype=ext4 rw console=ttyS0

OUTPUT kernel vmlinuz
OUTPUT initrd initrd.img
OUTPUT rootfs os.ext4

CHECK test -b /dev/vda
```

The kernel and the initrd are fetched for the check even if you build without
`-o`. See `examples/microvm`.

## How a check runs

1. miso builds the output.
2. miso boots it in a virtual machine with 2 CPUs and 2 GiB of memory. A disk
   is booted under UEFI. An ISO is booted as a CD. A rootfs is booted with
   its kernel and initrd.
3. When the system has booted, miso runs each `CHECK` of the output in the
   order of the file, each one in its own shell.
4. If every command exits with 0, the output is written under its name.

A disk is booted from a copy. Booting does not change the file that is
written.

The command is run by `/bin/sh` of the image. It is taken as written, so
pipes, `&&` and quotes work as in a shell. `ENV` values are not set in a
check.

Several checks after one output run in one boot. A check after a second
`OUTPUT` belongs to that one and boots it on its own.

Checks run during `miso build` with or without `-o`.

## When a check fails

The build stops. The output file is not written. The checks after the failing
one do not run.

```
miso: Imagefile:16: CHECK echo looking for emacs && command -v emacs: exit code 127
looking for emacs

to look at the layers before it: miso shell --before 16
```

The message names the line of the `CHECK` and the exit code of the command.
It shows what the command wrote. The last line is a `miso shell` command that
opens a shell on the layers before the check.

A passing check prints nothing.

If the system does not boot, the build fails with `the image did not boot`,
and with `within 5m0s` when it did not reach a booted state in five minutes.
The error names a log file in the cache directory. It holds the console
output of the boot, which shows where the boot stopped.

## What the machine needs

Checks run on the machine that runs miso, not in the builder VM. They need:

- `qemu-system-x86_64` and access to `/dev/kvm`
- the firmware, which miso downloads and keeps in the cache
- a vsock device. miso uses it to talk to the booted system. If there is
  none, the build stops with `CHECK needs a vsock of its own`.
