# list available recipes
default:
    @just --list

# build the static binary, the same file runs inside the builder vm
build:
    CGO_ENABLED=0 go build -o bin/miso ./cmd/miso

# test
test:
    go test ./...

# test with a coverage profile
cover:
    go test -coverprofile=coverage.out -covermode=atomic ./...

# lint
lint:
    golangci-lint run ./...

# format
fmt:
    golangci-lint fmt ./...

# check that package doc comments live in doc.go
doccheck:
    bash hack/check-doc-comments.sh

# check prose style: no em-dashes outside the license file
prose:
    bash hack/check-prose.sh

# check for known vulnerabilities in reachable code
vuln:
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# one-time dev setup after cloning: git hooks
setup: hooks

# install the git hooks
hooks:
    lefthook install

# everything that must pass before a push
ci: lint prose doccheck build test vuln
