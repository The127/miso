# miso

Dockerfile-style builds for systemd-based operating systems. A build file
of `FROM`, `COPY` and `RUN` steps becomes a stack of cached layers, and the
result comes out as standard systemd artifacts: disk images, ISOs,
discoverable disk images and the files systemd-sysupdate consumes. Every
image is booted and checked before it leaves the build.

miso builds images. What happens on the machine afterwards (applying
updates, configuration, first boot logic) is out of scope. miso provides
the data, you do it your way.

> **Status: early development.** Nothing here is usable yet.

## Usage

Not yet. The interface will be documented here once it exists.

## Contributing

Contributions are welcome, see [CONTRIBUTING.md](CONTRIBUTING.md). In
short: commits follow plain Conventional Commits (`type: description`, no
scopes) and must be DCO signed off (`git commit -s`). Both rules are
enforced by git hooks.

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

## Development setup

Required tooling:

- [Go](https://go.dev/) 1.27+.
- [golangci-lint](https://golangci-lint.run/) v2: linting, configured in
  `.golangci.yml`.
- [just](https://just.systems/): task runner. `just` lists the available
  recipes, `just ci` runs everything that must pass.
- [lefthook](https://lefthook.dev/): git hooks (lint, prose, doc comment
  and architecture checks on pre-commit, commit message and sign-off
  checks on commit-msg).

The package dependency rules in `arch-go.yml` are checked by
[arch-go](https://github.com/arch-go/arch-go), which the Go toolchain
fetches on its own (`go tool`). It is pinned in a module of its own,
`hack/tools/go.mod`, so its dependencies stay out of miso's.

After cloning, run the one-time setup. It activates the git hooks:

```bash
just setup
```

## License

miso is licensed under [AGPL-3.0-or-later](LICENSE).
