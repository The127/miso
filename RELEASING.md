# Releasing

Releases are made from the commit messages on `main`. Nobody tags by hand.

1. [release-please](https://github.com/googleapis/release-please) keeps one
   release pull request open. It holds the next version and the
   `CHANGELOG.md` entry, worked out from the Conventional Commits since the
   last release. `feat` bumps the minor version and `fix` the patch.
   Before 1.0.0, a breaking change also bumps only the minor version.
2. Merging that pull request makes the tag and the GitHub release.
3. In the same workflow, goreleaser builds `miso` for linux amd64 and
   arm64 and uploads a `tar.gz`, a `deb` and an `rpm` for each, and
   `checksums.txt`, to that release. Tags with a suffix such as `-rc.1`
   are marked as pre-releases.
4. Also in that workflow, cosign signs `checksums.txt` without a key. The
   workflow's own identity signs it, so there is no key or secret to keep.
   The signature is uploaded as `checksums.txt.sigstore.json`.

The binary reports its version from the Go build info stamp. A clean
checkout of the tag is enough, there are no ldflags.

## What the packages hold

Only the `miso` binary and the license. The builder kernel and the UEFI
firmware are not shipped. The first build fetches each by its pinned
digest. The packages depend on the qemu of their arch:

| Arch | deb | rpm |
|---|---|---|
| amd64 | `qemu-system-x86` | `qemu-system-x86-core` |
| arm64 | `qemu-system-arm` | `qemu-system-aarch64-core` |

## One-time setup

The release pull request and the tag must start other workflows, which
the default `GITHUB_TOKEN` cannot do. So the workflow uses a GitHub App:

1. Create a GitHub App with read and write access to contents and pull
   requests, and install it on this repository.
2. Add its ID as the repository variable `RELEASE_APP_ID`.
3. Add its private key as the repository secret `RELEASE_APP_PRIVATE_KEY`.

## Forcing a version

To release a version the commits would not give, add `"release-as"` with
that version to `release-please-config.json`. Remove it again once that
release is out, or the same version is proposed every time.

## Checking the configuration

```
just release-check
just release-snapshot
```

The first validates `.goreleaser.yaml`. The second builds every artifact
into `dist/` without publishing anything.
