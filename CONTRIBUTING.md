# Contributing to miso

## Development setup

Required tooling:

- [Go](https://go.dev/) 1.27+.
- [golangci-lint](https://golangci-lint.run/) v2: linting, configured in
  `.golangci.yml`.
- [bats](https://bats-core.readthedocs.io/): tests of the built `miso`
  from the outside, run by `just test-cli`.
- [mdBook](https://rust-lang.github.io/mdBook/): builds the documentation
  site in `site/`, run by `just docs`.
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

## Developer Certificate of Origin

Contributions must be signed off. By adding a `Signed-off-by` line to your
commit you certify the [Developer Certificate of Origin 1.1](https://developercertificate.org/):
that you wrote the contribution or otherwise have the right to submit it
under the project's license.

Sign off with git's built-in flag:

```
git commit -s
```

which appends a line of the form

```
Signed-off-by: Your Name <your@email.example>
```

Commits without a sign-off are rejected by the commit-msg hook (and will be
rejected in review).

## Commit messages

Plain [Conventional Commits](https://www.conventionalcommits.org/): a first
line of the form `type: description`, no scopes. Allowed types: `feat`,
`fix`, `chore`, `docs`, `refactor`, `test`, `perf`, `build`, `ci`, `revert`.

## AI-assisted contributions

Welcome, under three rules. Disclose AI-generated code with a co-author
trailer (e.g. `Co-Authored-By: Claude <noreply@anthropic.com>`). You remain
fully responsible for what you submit: correctness, license compatibility,
and your DCO sign-off. And you must have reviewed and understood the code
yourself before pushing it.

## Licensing of contributions

miso is AGPL-3.0-or-later. Every contribution is licensed under AGPL-3.0-or-later.
