# Outputs

An `OUTPUT` instruction writes a file from a stage. This page lists every
kind of output with what it contains, what the image needs for it and how
to use the result. The syntax of `OUTPUT` is on the
[Imagefile](imagefile.md) page.

| Kind | Result | Booted by a `CHECK` |
| --- | --- | --- |
| [`disk`](#disk) | A bootable GPT disk image. | yes |
| [`iso`](#iso) | A bootable ISO image. | yes |
| [`rootfs`](#rootfs) | The root file system as an image. | yes |
| [`kernel`](#kernel-and-initrd) | The kernel of the image. | no |
| [`initrd`](#kernel-and-initrd) | The initrd of the image. | no |
| [`portable`](#portable) | A portable service. | no |
| [`sysext`](#sysext) | A system extension. | no |
| [`confext`](#confext) | A configuration extension. | no |
| [`update`](#update) | Files for `systemd-sysupdate`. | no |

The output files are written to the directory given with `miso build -o`.

The kernel and the initrd of an image are found in the image. miso takes the
newest version under `/usr/lib/modules`.

## disk

```
OUTPUT disk os.raw
```

A GPT disk image that boots under UEFI. The layout comes from the
`PARTITION` lines above the `OUTPUT`. The disk is made by `systemd-repart`.

miso builds a unified kernel image from the kernel, the initrd and the
`CMDLINE` of the stage, and puts it on the EFI system partition together with
systemd-boot. The image therefore needs:

- a kernel and an initrd
- `systemd-boot` and its stub, which the Debian package `systemd-boot-efi`
  provides

`examples/htop` builds a disk with an EFI system partition and a root
partition. The [quick start](quick-start.md) walks through it.

A disk is written as a sparse file.

## iso

```
OUTPUT iso os.iso
```

The same disk with a boot catalog, so that it boots from a CD drive. It is made from the same `PARTITION` and `CMDLINE` lines.

A root on a CD is not found the way a root on a disk is. The initrd needs the
drivers of the CD and a way to find a root there. `examples/iso` makes the
initrd again with `dracut`, with the loop unit and `systemd-dissect` added.
It uses a read-only erofs root with an overlay in memory on top.

A `CHECK` boots the ISO as a CD.

## rootfs

```
OUTPUT rootfs os.ext4
OUTPUT rootfs os.erofs --format=erofs
```

The root file system as an image. `--format` is `ext4`, which is the default,
or `erofs`.

A rootfs does not contain a kernel. Write `OUTPUT kernel` and `OUTPUT initrd`
next to it to get the files to boot it with. `examples/microvm` does this and
the result boots in Firecracker and in QEMU.

A `CHECK` on a rootfs boots the kernel and the initrd of the stage with the
rootfs as a virtual disk. So the stage needs, above the rootfs:

- an `OUTPUT kernel`
- an `OUTPUT initrd`
- a `CMDLINE` with a `root=` word, such as `root=/dev/vda`

The initrd must be able to mount the root. Debian's cloud initrd is made for
the disk of the cloud image, so `examples/microvm` makes it again with the
driver of the virtual disk.

## kernel and initrd

```
OUTPUT kernel vmlinuz
OUTPUT kernel vmlinux --elf
OUTPUT initrd initrd.img
```

The kernel and the initrd of the image, copied as they are.

`--elf` unpacks the kernel to the ELF file it contains. Firecracker needs
that form.

These outputs cannot be booted, so they take no `CHECK`.

## portable

```
OUTPUT portable hello.raw
OUTPUT portable hello.raw --format=erofs
```

A portable service: the root file system in a disk image with one root
partition. It does not boot. The image needs `/usr/lib/os-release` or
`/etc/os-release`.

On a machine with systemd:

```
portablectl attach ./hello.raw
systemctl start hello
```

`--format` is `ext4` or `erofs`. `examples/portable` builds a service with
busybox.

## sysext

```
OUTPUT sysext hello.raw
```

A system extension: files for `/usr` and `/opt` in a disk image with one root
partition. The stage should hold only what the extension adds, so it usually
starts `FROM scratch` and copies in what it needs.

The stage needs a release file named after the output file without `.raw`:

```
/usr/lib/extension-release.d/extension-release.hello
```

With `ID=_any` in it, the extension merges on any operating system. A real
extension names the system it was made for.

On a machine with systemd:

```
cp hello.raw /var/lib/extensions/
systemd-sysext merge
```

`--format` is `ext4` or `erofs`. `examples/sysext` builds one.

## confext

```
OUTPUT confext hello.raw
```

A configuration extension: files for `/etc` in a disk image with one root
partition. It is made like a sysext. The release file is

```
/etc/extension-release.d/extension-release.hello
```

On a machine with systemd:

```
cp hello.raw /var/lib/confexts/
systemd-confext merge
```

`--format` is `ext4` or `erofs`. `examples/confext` builds one.

## update

```
OUTPUT update updates --version=1.0
```

The files that `systemd-sysupdate` downloads, made from the partitions of the
disk. `--version` is required. It may use letters, digits and `. _ + ^ ~ -`,
and must start with a letter or a digit.

The output is a directory. Name the partitions to ship with `SplitName=` in
their `PARTITION` line. The directory contains:

- `<SplitName>_<version>.raw` for every partition that has a `SplitName`
- `uki_<version>.efi`, the unified kernel image of the disk
- `SHA256SUMS`, the digests of those files

For `SplitName=root` and `--version=1.0`, the files are `root_1.0.raw`,
`uki_1.0.efi` and `SHA256SUMS`.

An update is an error when no `PARTITION` above it has a `SplitName`.

Serve the directory over HTTP, or point a transfer file at it with a
`file://` URL. `examples/update` has two transfer files that install the
update into plain files. A target that is a partition was not tried.

## Errors

An error in an `OUTPUT` names the line, the kind and the file name, for
example `Imagefile:22: OUTPUT iso os.iso: the image did not boot within 5m0s`.
It also gives the reason. These are the ones from the `OUTPUT` line itself:

- `unknown kind of output`
- `--x: unknown option`
- `--elf: option takes no value`
- `--format=x: unknown file system`
- `not a plain file name`, when the name has a `/` in it
- `file name taken`, when two outputs have the same name
