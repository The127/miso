# miso

Dockerfile-style builder for systemd-based operating systems. A build file
becomes cached layers inside a builder VM, and the result comes out as
standard systemd artifacts: disk images, ISOs, discoverable disk images and
systemd-sysupdate files. Every image is booted and checked before it leaves
the build. `cmd/miso` is the binary.

Code that runs on the target after installation (applying updates,
configuration, first boot logic) is out of scope. miso emits data in
standard systemd formats and leaves the rest to whoever runs the machine.

## Working agreements

- Commit messages follow plain Conventional Commits: `type: description`
  (`feat`, `fix`, `chore`, `docs`, `refactor`, `test`, ...). No scopes.
- Every commit is DCO signed off: always `git commit -s`. Enforced by the
  commit-msg hook.
- Work incrementally in small reviewed steps, no big-bang generation.
- This is the tool itself, not a first cut of it. Never frame plans as
  "v1", versions or phases, and never propose cutting scope to ship sooner.
  Talk about the order of work instead.
- Keep comments short. Never explain what the code does, the code is the
  documentation (it cannot go stale). Comment only the why, and only when
  it is non-obvious.
- No em-dashes and no semicolons in prose: documents, comments, and commit
  messages. Semicolons in code only where Go syntax requires them.
- Package doc comments live in a dedicated `doc.go` per package, never on
  another file's package clause.
- Every package is covered by a dependency rule in `arch-go.yml`. A new
  package arrives together with its rule.
- Use the justfile: `just ci` must pass before a commit. `just setup`
  prepares a fresh clone.

## Go tests

- External test packages (`package foo_test`), testing the exported
  surface. Internals reach tests through `export_test.go` when needed.
- testify only: `require` for preconditions, `assert` for the checks. No
  table-driven tests, no `t.Run` subtests.
- One behaviour per test, named as a sentence:
  `TestParseErrorNamesTheDocument`.
- Test bodies are split by `// arrange`, `// act`, `// assert` comments.
- Fixture helpers take `t.Helper()`. Use `t.TempDir()` and `t.Setenv`.

## AI usage rules

These are enforced for all AI-assisted work in this repo (see README /
CONTRIBUTING for the contributor-facing policy):

- Every commit containing AI-generated code carries a co-author trailer,
  e.g. `Co-Authored-By: Claude <noreply@anthropic.com>`.
- Commits do NOT carry session links or other tool-run metadata (no
  `Claude-Session:` trailers). This repo overrides that default.
- The human contributor stays responsible: AI agents propose changes and
  commit only after the contributor has reviewed, understood, and approved
  them. The gate is at the commit, not the push.

## Licensing

The whole repository is a single Go module licensed under AGPL-3.0-or-later.

## Personal extensions

Machine- or person-specific notes go in `CLAUDE.local.md` (untracked):

@CLAUDE.local.md
