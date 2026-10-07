# The Imagefile

An Imagefile describes how to build an image. This page lists every
instruction and the rules of the file. `miso build` and `miso plan` read a
file named `Imagefile` in the directory you give them, or the file named by
`-f`.

Example Imagefile:
```
FROM debian:sid
RUN apt-get update && apt-get install -y htop
OUTPUT rootfs os.ext4
```

## File format

- One instruction per line. The keyword comes first. Keywords are not case
  sensitive, so `run` and `RUN` are the same.
- Spaces and tabs before a keyword are ignored.
- Blank lines are ignored.
- A line whose first character is `#` is a comment. A `#` later in a line is
  not a comment. It is part of the instruction.
- A line that ends in `\` continues on the next line. The backslash is
  removed and the lines are joined without a newline. The next line's leading
  spaces stay.
- A file needs at least one `FROM`. Every other instruction must come after
  one.

An error names the line where the instruction starts, for example
`Imagefile:4: COPY needs an absolute destination`.

## Stages

Each `FROM` starts a stage. The instructions after it belong to that stage
until the next `FROM`. A stage is a root file system that the instructions
build up step by step.

Instructions fall into three groups:

- `RUN`, `ENV` and `COPY` change the root file system.
- `PARTITION` and `CMDLINE` describe a disk and leave the root file system
  alone.
- `OUTPUT` and `CHECK` make files and leave the root file system alone.

An `OUTPUT` takes the root file system as it is at that line. Instructions
below it do not change what it captured. The `PARTITION` and `CMDLINE` lines
above an `OUTPUT` in the same stage belong to it, and the ones below do not.

A stage can have any number of outputs. A stage without outputs is useful as
a base for another stage or as the source of a `COPY --from`.

## FROM

```
FROM <base>
FROM <base> AS <name>
```

Starts a stage from a base. The base is one of:

- `scratch`, an empty root file system.
- The name of an earlier stage. The new stage starts where that stage ends,
  including its `ENV` values.
- `debian:sid`, `debian:13` or `debian:trixie`. These are the Debian cloud
  images for the architecture miso runs on. `debian:13` and `debian:trixie`
  are the same image. A name means whatever image is current at its address
  when it is first fetched. After that the fetched image is cached.

Any other name is an error. Paths, URLs and container images are not yet
supported.

`AS <name>` names the stage so that a later stage can start `FROM <name>` or
use `COPY --from=<name>`. Names are case sensitive and unique in the file.
`scratch` cannot be a stage name.

## RUN

```
RUN [--network=none|default] <command>
```

Runs a shell command in the stage's root file system. The command is run
with `/bin/sh -c`, so the image needs a `/bin/sh`. The rest of the line is
the command, exactly as written.

Every `RUN` runs as root in the builder VM, in its own process, mount and
network namespaces, with `/` as the working directory. The environment is
`PATH`, `HOME=/root` and the values set by `ENV`. The environment of the
host is not passed on.

A command that exits with a code other than 0 fails the build.

`--network=none` gives the command no network except loopback. The default
is `--network=default`, which gives it network access. Giving `default` is
the same as giving no option.

The result of a `RUN` is cached as a layer. The cache key depends on the
command and on every step before it.

## ENV

```
ENV <key>=<value> [<key>=<value> ...]
```

Sets environment variables for the `RUN` instructions after it in the same
stage, and for stages that start `FROM` this stage. A later `ENV` with the
same key replaces the earlier value.

A value with spaces is put in double quotes:

```
ENV GREETING="hello world" LEVEL=2
```

A value is everything after the first `=`. Single quotes and backslash
escapes are not special.

`ENV` does not write anything into the image. It does not reach `COPY` and
it does not reach `CHECK`. A variable that the booted system needs has to be
put into the image by a `RUN`.

## COPY

```
COPY <source>... <destination>
COPY --from=<stage> <source>... <destination>
```

Copies files into the root file system.

The destination must be an absolute path. With more than one source it must
end in `/`. Paths with spaces are put in double quotes.

### From the build context

Without `--from`, the sources are paths relative to the build context, the
directory you pass to `miso build`.

- A source is a literal path. Patterns such as `*` are not expanded.
- A source must exist and must lie inside the context. A path with `..` or
  a symbolic link in the middle of it is an error. A symbolic link as the last
  part is copied as a link.
- A directory source copies its contents into the destination. `COPY app
  /opt/app/` puts the files of `app` directly in `/opt/app/`.
- A file goes into the destination if the destination ends in `/` or is a
  directory that already exists. Otherwise the destination is the new name of
  the file.
- Missing parent directories are created with mode 0755. Files that already
  exist are replaced.
- Everything is owned by root. The mode of a file is its mode in the context.
  Times are not kept.

Changing a file in the context changes the cache key of the `COPY` and of
every step after it.

### From another stage

With `--from=<stage>`, the sources are absolute paths in the root file
system of an earlier stage, as it is at the end of that stage:

```
FROM debian:sid AS program
RUN apt-get update && apt-get install -y busybox-static

FROM scratch
COPY --from=program /bin/busybox /usr/bin/hello
```

Owner, mode and extended attributes are kept. The destination rules are the
same as above.

Copying a file made by an `OUTPUT` of another stage is not supported yet.

## CMDLINE

```
CMDLINE <text>
```

Adds to the kernel command line of the disks of the stage. The text is taken
as written. Several `CMDLINE` lines are joined with one space, in file order.

`CMDLINE` is used by `OUTPUT disk`, `OUTPUT iso` and `OUTPUT update`. It
becomes the command line of the unified kernel image on the disk and takes
the place of `/etc/kernel/cmdline` in the image. Without a `CMDLINE`, the
image's own `/etc/kernel/cmdline` is used if it has one.

A `CHECK` on an `OUTPUT rootfs` boots the kernel with this command line.

Only the lines above an `OUTPUT` count for it. Lines are not passed on to a
stage that starts `FROM` this stage. The other kinds of output ignore
`CMDLINE`.

## PARTITION

```
PARTITION <name> [<Key>=<Value> ...]
```

Adds a partition to the disks of the stage. The settings are the ones
`systemd-repart` takes in a partition definition file, such as `Type`,
`Format`, `CopyFiles`, `SizeMinBytes` and `SplitName`. See
`repart.d(5)` for the list.

```
PARTITION esp Type=esp Format=vfat CopyFiles=/efi:/ SizeMinBytes=256M SizeMaxBytes=256M
PARTITION root Type=root Format=ext4 CopyFiles=/:/ SizeMinBytes=3G
```

- Settings are separated by spaces. A value cannot contain a space.
- A key can appear more than once, as `CopyFiles` often does.
- miso does not check the settings. `systemd-repart` reports a bad one
  when the disk is made.
- The name is only used for the file name of the definition. It has to be a
  plain file name, and it is unique within a stage.
- The partitions are created in the order of the lines.

With at least one `PARTITION`, only those partitions are used. With none,
`systemd-repart` reads the definitions it finds in the image.

`PARTITION` is used by `OUTPUT disk`, `OUTPUT iso` and `OUTPUT update`. For
the other kinds of output it does nothing, and miso does not warn about it.

Only the lines above an `OUTPUT` count for it.

The disk needs `systemd-boot` and its stub in the image, so install
`systemd-boot-efi`, and it needs a kernel and an initrd.

## OUTPUT

```
OUTPUT <kind> <name> [--option[=value] ...]
```

Writes a file from the stage. The name is a plain file name and is unique in
the whole file, over all stages. The files are written to the directory given
to `miso build -o`.

| Kind | Result | Options |
| --- | --- | --- |
| `disk` | A GPT disk image with a unified kernel image. | |
| `iso` | The same disk with a boot catalog for optical media. | |
| `rootfs` | The root file system as an image. | `--format=ext4` (default) or `--format=erofs` |
| `portable` | A portable service in a disk image with one root partition. | `--format` |
| `sysext` | A system extension in a disk image with one root partition. | `--format` |
| `confext` | A configuration extension in a disk image with one root partition. | `--format` |
| `update` | A directory with the files `systemd-sysupdate` uses. | `--version=<version>` (required) |
| `kernel` | The kernel of the image. | `--elf` unpacks an x86 kernel to an ELF file, `--path=<file>` names it |
| `initrd` | The initrd of the image. | `--path=<file>` names it |

The kernel and the initrd are found in the image. miso uses the newest
version under `/usr/lib/modules`, unless `--path` names the file.

Disks are written as sparse files.

miso adds a stage named `miso tools` to the build when an output needs
tools to be made. It holds `systemd-repart`, `ukify`, the file system tools
and compressors, installed from a fixed Debian snapshot. Plain `kernel` and
`initrd` outputs do not need it.

## CHECK

```
CHECK <command>
```

Boots the output above it and runs the command in the booted system. The
command is run by `/bin/sh` in the image. It is taken as written.

- A `CHECK` belongs to the closest `OUTPUT` above it in the same stage.
  Several checks after one output run in order, in one boot.
- The output has to be a `disk`, an `iso` or a `rootfs`. The other kinds are
  never booted, and a `CHECK` for them is an error.
- A `rootfs` needs a kernel and an initrd to boot. Put an `OUTPUT kernel` and
  an `OUTPUT initrd` above it. It also needs a `CMDLINE` with a `root=`
  word above it.
- The image needs systemd.
- `ENV` values are not set.

If a command exits with a code other than 0, the build fails at that line
and the output is not written. Checks after it do not run.

A check runs during `miso build`, even without `-o`, for disks and for the
kernel and initrd that a checked `rootfs` is booted with.
