# Command line

```
miso <command> [options] [arguments]
```

| Command | What it does |
| --- | --- |
| [`plan`](#miso-plan) | Lists the steps of a build without building. |
| [`build`](#miso-build) | Builds the outputs of an Imagefile. |
| [`shell`](#miso-shell) | Opens a shell on the layers at a step of the build. |
| [`prune`](#miso-prune) | Removes old layers and downloads from the cache. |
| [`resize`](#miso-resize) | Makes the cache disk larger. |
| [`version`](#miso-version) | Prints the version. |

`miso --help` and `miso <command> --help` print the options of a command.

## The cache

miso keeps everything it fetches and builds in a `miso` directory inside the
user cache directory. On Linux that is `$XDG_CACHE_HOME/miso`, or
`~/.cache/miso` if the variable is not set.

- `bases/` holds the downloaded base images.
- `builder/layers.img` is the disk the builder VM keeps the layers on. It is
  a sparse file. It takes only the space the layers use and can grow up to
  50 GiB. `miso resize` makes it larger.

Only one miso uses the cache disk at a time. A second one waits and prints
`miso: waiting for another build to let go of the cache`.

## The build context and the build file

`plan`, `build` and `shell` read an Imagefile. They take the build context as
an optional argument. The context is a directory and defaults to the current
one. The build file is the file named `Imagefile` in the context. `-f` names
another file. A path given to `-f` is used as it is and is not relative to the
context. `COPY` always reads from the context.

## miso plan

```
miso plan [-f <file>] [context]
```

Reads the Imagefile and lists what a build would do, without building.

The first line names the version of miso. A base image that is not in the
cache yet is listed as `download <name>`. Then every stage is listed with its
`FROM` line, and under it every step with its line number, its cache key,
whether its layer is cached, and the step as written. A stage named
`miso tools` is added by miso when an output needs the tools to be made.

After the key, `RUN`, `COPY` and `OUTPUT` say `cached` when the layer is on
the cache disk and `run` when a build would make it. With no cache disk, or a
base image that is not downloaded yet, every one of them says `run`. To know
what is cached, `miso plan` starts the builder VM, so it waits while another
build uses the cache. It does not start the VM when there is nothing to ask.

Two plans with the same cache key for a step build the same layer. After a
change, the keys of the changed step and of every step after it are different.

An error in the Imagefile is reported with the file name and the line, for
example `Imagefile:4: COPY needs an absolute destination`.

## miso build

```
miso build [-f <file>] [-o <directory>] [context]
```

Builds the Imagefile in the builder VM. The steps are run one by one and
their output is printed. A step whose cache key is already in the cache is not
run again.

| Option | Meaning |
| --- | --- |
| `-f`, `--file` | The build file. `Imagefile` in the context if not given. |
| `-o`, `--output` | The directory the outputs are written to. It is created if it is missing. |

Without `-o` no output is written. Disks with a `CHECK` are still built and
booted, so the checks run.

The build stops at the first step that fails. An error in a step names the
line of the Imagefile, and for a step that is not a `FROM` it also prints a
`miso shell` command that opens a shell before that step.

The builder VM needs access to `/dev/kvm`.

## miso shell

```
miso shell [-f <file>] [--before <line>] [context] [-- <command>...]
```

Opens an interactive shell in the root file system as it is at some point of
the build. Layers that are not built yet are built first.

| Option | Meaning |
| --- | --- |
| `-f`, `--file` | The build file. `Imagefile` in the context if not given. |
| `--before` | The line of the Imagefile whose step the shell stands before. |

Without `--before` the shell stands at the end of the last stage. With
`--before 12` it stands before the step on line 12, with the environment
that step would have and the network it would have. A line that holds no step
is an error. A `FROM` line is not a step.

Everything written in the shell is thrown away when it ends. It does not
change the cache.

Words after `--` are run instead of the interactive shell:

```
miso shell --before 12 -- ls /etc
```

The exit code of the shell, or of the command, is the exit code of `miso`.

## miso prune

```
miso prune [--older-than <duration>] [--keep-storage <size>]
```

Removes layers and downloaded files that were not used for a while. At
least one option is needed. Without one, `prune` stops with an error.

| Option | Meaning |
| --- | --- |
| `--older-than` | Remove what no build used for longer than this. A duration such as `720h`. |
| `--keep-storage` | Remove the least recently used until the layers take no more than this, and the downloads take no more than this. A size such as `10GiB`. |

Both limits must be above zero. With both, what either of them selects is
removed.

`prune` prints how many downloads it removed and how many bytes that freed.

## miso resize

```
miso resize [--size <size>]
```

Makes the cache disk larger. The layers on it are kept. It can only grow the
disk. A size smaller than the disk is an error and leaves the disk as it was.

| Option | Meaning |
| --- | --- |
| `--size` | The new size, such as `80GiB`. 50 GiB if not given. |

Without `--size`, a disk that already has 50 GiB or more is left alone.

`resize` needs an existing cache disk. It does not make one. The first build
makes it.

Sizes are written as a number and a unit, such as `500MB` or `10GiB`.

## miso version

```
miso version
```

Prints the version of this build of miso.
