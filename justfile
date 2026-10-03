# list available recipes
default:
    @just --list

# build the static binary, the same file runs inside the builder vm
build:
    CGO_ENABLED=0 go build -o bin/miso ./cmd/miso

# test
test:
    go test -race ./...

# test what needs root and a real kernel, inside vms (needs /dev/kvm), on
# the host's kernel and on the base image's, all at once
test-vm *args:
    bash hack/test-vm-kernels.sh {{args}}

# test one vm package on the host's kernel, stopping at the first failure,
# for the red and green of one test, for example
# just test-vm-one internal/agent -test.run=TestName
test-vm-one package *args:
    MISO_VMTEST_PACKAGES={{package}} bash hack/test-vm.sh -test.timeout=20s -test.failfast {{args}}

# boot the builder kernel in a real VM through the qemu driver, with the
# static test binary or the miso binary as the VM's init. Needs /dev/kvm,
# /dev/vhost-vsock, qemu-system-x86_64 and, the first time, the network
# for the kernel
test-kvm *args: build
    MISO={{justfile_directory()}}/bin/miso CGO_ENABLED=0 go test -tags kvm -count=1 ./internal/qemu/ ./internal/check/ ./cmd/miso/ -run 'BuilderKernel|WithKVM' {{args}}

# test the miso binary from the outside, as a user runs it
test-cli: build
    MISO={{justfile_directory()}}/bin/miso bats test

# test with a coverage profile
cover:
    go test -race -coverprofile=coverage.out -covermode=atomic ./...

# lint
lint:
    golangci-lint run ./...

# format
fmt:
    golangci-lint fmt ./...

# check the package dependency rules in arch-go.yml
arch:
    go tool -modfile=hack/tools/go.mod arch-go

# describe the package dependency rules in prose
arch-describe:
    go tool -modfile=hack/tools/go.mod arch-go describe

# check that package doc comments live in doc.go
doccheck:
    bash hack/check-doc-comments.sh

# check prose style: no em-dashes outside the license file
prose:
    bash hack/check-prose.sh

# validate the goreleaser config
release-check:
    go run github.com/goreleaser/goreleaser/v2@latest check

# build the release artifacts locally without publishing
release-snapshot:
    go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean

# check for known vulnerabilities in reachable code
vuln:
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# one-time dev setup after cloning: git hooks
setup: hooks

# install the git hooks
hooks:
    lefthook install

# everything that must pass before a push
ci: lint prose doccheck arch build cover test-cli vuln
