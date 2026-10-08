# miso

[![ci](https://github.com/The127/miso/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/The127/miso/actions/workflows/ci.yml)
[![coverage](https://codecov.io/gh/The127/miso/graph/badge.svg)](https://app.codecov.io/gh/The127/miso)

miso builds operating system images from a build file that looks like
a Dockerfile. The steps are cached as layers. The result comes out as
standard systemd artifacts: disk images, ISOs, root file systems, portable
services, system extensions, configuration extensions and the files
`systemd-sysupdate` uses. A `CHECK` step boots the result and runs a command
in it before the build writes anything.

miso builds images. It does not apply updates, configure machines or run
anything on the target after installation.

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

```
miso build -o out
```

This builds a Debian disk with htop in it, boots it, and writes `out/os.raw`.

## Documentation

The documentation is at <https://the127.github.io/miso>. It has a quick
start, the reference for the Imagefile and the command line, and examples for
every kind of output. The same examples are in the [examples](examples)
directory.

## Installing

Each [release](https://github.com/The127/miso/releases) has a `tar.gz`,
a `deb` and an `rpm` for linux amd64 and arm64, with `checksums.txt`.
The packages depend on qemu. They do not hold the builder kernel or the
firmware, so the first build downloads them and checks their digests.

miso needs access to `/dev/kvm`.

## Contributing

Contributions are welcome, see [CONTRIBUTING.md](CONTRIBUTING.md), which also
has the development setup. In short: commits follow plain Conventional Commits
(`type: description`, no scopes) and must be DCO signed off (`git commit -s`).
Both rules are enforced by git hooks.

## AI-assisted contributions

AI-assisted contributions need to follow these rules:

- **Disclose it.** Commits with AI-generated code carry a co-author
  trailer, e.g. `Co-Authored-By: Claude <noreply@anthropic.com>`.
- **You are the author.** The contributor is fully responsible for
  AI-generated code: its correctness, its license compatibility, and the
  DCO sign-off certifying the right to submit it. "The AI wrote it" is
  never an excuse.
- **Review it yourself, before pushing.** Submit only code you have read,
  understood, and could explain and defend in review as your own.

## Security

Please report vulnerabilities privately, see [SECURITY.md](SECURITY.md).

## License

miso is licensed under [AGPL-3.0-or-later](LICENSE).
