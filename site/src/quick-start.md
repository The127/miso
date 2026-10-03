# Quick start

This page builds a Debian disk image with htop in it.

You need Linux on amd64, `qemu-system-x86_64` and access to `/dev/kvm`.
The first build downloads the builder kernel, the firmware and the Debian
base image.

## The Imagefile

Make a directory and put a file named `Imagefile` in it:

```
FROM debian:sid
RUN apt-get update && apt-get install -y --no-install-recommends systemd-boot-efi htop

RUN : > /etc/fstab

CMDLINE rw console=ttyS0

PARTITION esp Type=esp Format=vfat CopyFiles=/efi:/ SizeMinBytes=256M SizeMaxBytes=256M
PARTITION root Type=root Format=ext4 CopyFiles=/:/ SizeMinBytes=3G

OUTPUT disk os.raw

CHECK command -v htop
```

This is `examples/htop/Imagefile` from the repository without its comments.

| Line | Meaning |
| --- | --- |
| `FROM debian:sid` | Start from the Debian sid cloud image. |
| `RUN apt-get ...` | Install htop and the systemd-boot files the disk needs. |
| `RUN : > /etc/fstab` | Empty the fstab. The cloud image's fstab names partitions the new disk does not have. |
| `CMDLINE` | The kernel command line of the disk. |
| `PARTITION` | A partition of the disk, written as `systemd-repart` settings. |
| `OUTPUT disk os.raw` | Write the disk to a file named `os.raw`. |
| `CHECK` | Boot the disk and run this command in it. |

## Plan

`miso plan` lists the steps without building anything. Run it in the
directory with the Imagefile:

```
miso plan
```

Each stage starts with its `FROM` line. Under it, every step shows its line
number, its cache key, whether the layer is `cached` or would `run`, and the
step as written. The stage named `miso tools` is not in your file. miso adds
it for the tools that write the disk.

## Build

```
miso build -o out
```

miso prints the output of every step. When the build succeeds, the disk is
`out/os.raw`. It is a GPT disk with an EFI system partition and a root
partition, and it boots under UEFI.

## Build again

miso keeps the result of every step as a layer. If you change a step, miso
builds that step and all steps after it again and reuses the rest. Edit the
`RUN` line with htop and run `miso plan` to see it: the cache key of that
step and of every step after it changes, and those steps say `run`.
