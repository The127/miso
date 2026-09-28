#!/bin/bash
set -euo pipefail

if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

cd "${CLAUDE_PROJECT_DIR:-$(git rev-parse --show-toplevel)}"

# Versions come from the files CI reads, so the hook cannot drift from them
golangci_version=$(sed -nE 's/^ *version: (v[0-9.]+)$/\1/p' .github/workflows/ci.yml | head -n1)
go_minor=$(sed -nE 's/^go ([0-9]+\.[0-9]+).*$/\1/p' go.mod)
if [ -z "$golangci_version" ] || [ -z "$go_minor" ]; then
  echo "cannot read golangci-lint or go version from ci.yml and go.mod" >&2
  exit 1
fi

apt_missing=()
for tool in just bats; do
  command -v "$tool" >/dev/null || apt_missing+=("$tool")
done
if [ "${#apt_missing[@]}" -gt 0 ]; then
  apt-get update -qq
  apt-get install -y "${apt_missing[@]}"
fi

gopath_bin="$(go env GOPATH)/bin"
export PATH="$gopath_bin:$PATH"

if ! command -v lefthook >/dev/null; then
  go install github.com/evilmartians/lefthook@latest
fi

# An older build refuses the module, so build with the Go that go.mod names
if ! golangci-lint --version 2>/dev/null | grep -q "has version ${golangci_version#v} built with go${go_minor}"; then
  GOTOOLCHAIN="go${go_minor}.0" go install \
    "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${golangci_version}"
fi

if [ -n "${CLAUDE_ENV_FILE:-}" ]; then
  echo "export PATH=\"$gopath_bin:\$PATH\"" >> "$CLAUDE_ENV_FILE"
  # `just vuln` runs govulncheck@latest, which would otherwise pick an older toolchain than go.mod needs
  echo "export GOTOOLCHAIN=go${go_minor}.0" >> "$CLAUDE_ENV_FILE"
fi

just setup
