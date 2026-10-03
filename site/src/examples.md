# Examples

The `examples` directory of the repository has one build for each kind of
output. Each directory holds an `Imagefile` with comments, and some hold files
that it copies in. Build one from the root of the repository:

```
miso build -o out examples/<name>
```

All of them start from `debian:sid` and build with the steps described on the
[Imagefile](imagefile.md) page.

| Example | What it builds |
| --- | --- |
| [`htop`](#htop) | A bootable disk with htop. |
| [`iso`](#iso) | A bootable ISO. |
| [`microvm`](#microvm) | A root file system with a kernel and an initrd. |
| [`portable`](#portable) | A portable service. |
| [`sysext`](#sysext) | A system extension. |
| [`confext`](#confext) | A configuration extension. |
| [`update`](#update) | The files of an update for `systemd-sysupdate`. |

## htop

A Debian disk with htop in it, with an EFI system partition and a root
partition. The `CHECK` boots the disk and looks for htop. The
[quick start](quick-start.md) builds it.

Result: `out/os.raw`

## iso

A Debian ISO that boots from a CD. The root is a read-only erofs file system
with an overlay in memory on top. The initrd is made again with `dracut`, so
that it can find a root on a CD. The `CHECK` boots the ISO as a CD.

Result: `out/os.iso`

## microvm

A Debian root file system for a microVM, with the kernel and the initrd to
boot it. The initrd is made again with the driver of the virtual disk. The
`CHECK` boots the three files together.

Result: `out/os.ext4`, `out/vmlinuz` and `out/initrd.img`

The `Imagefile` has a `qemu-system-x86_64` command that boots them. Firecracker
needs the kernel as an ELF file, so write `OUTPUT kernel vmlinux --elf`
instead of `OUTPUT kernel vmlinuz`. The `Imagefile` also has a Firecracker
configuration that boots the result to a login prompt.

## portable

A portable service. It holds busybox and a unit named `hello` that prints a
line and sleeps.

Result: `out/hello.raw`

On a machine with systemd:

```
portablectl attach ./out/hello.raw
systemctl start hello
```

## sysext

A system extension that adds one program, `/usr/bin/hello`. The stage starts
`FROM scratch` and copies busybox out of a stage that installs it, so the
image holds only what the extension adds. The release file is copied in from
the example directory.

Result: `out/hello.raw`

On a machine with systemd:

```
cp out/hello.raw /var/lib/extensions/
systemd-sysext merge
hello
```

## confext

A configuration extension that adds `/etc/hello.conf`. It starts `FROM
scratch` and copies in the file and the release file.

Result: `out/hello.raw`

On a machine with systemd:

```
cp out/hello.raw /var/lib/confexts/
systemd-confext merge
cat /etc/hello.conf
```

## update

The files of an update for `systemd-sysupdate`, made from the partitions of
a disk. The root partition has `SplitName=root`, so it is shipped.

Result: the directory `out/updates` with `root_1.0.raw`, `uki_1.0.efi` and
`SHA256SUMS`.

The `Imagefile` has two transfer files that install the update into plain
files. They read the update from a `file://` URL.
