#!/usr/bin/env bats
# MISO names the miso binary under test, see just test-cli.

own_arch() {
    case $(uname -m) in
    x86_64) echo amd64 ;;
    aarch64 | arm64) echo arm64 ;;
    *) uname -m ;;
    esac
}

@test "a build for an arch it does not build for fails and names the flag and the arch it builds for" {
    run "$MISO" build --arch sparc

    [ "$status" -eq 1 ]
    [[ "$output" == *'--arch "sparc"'* ]]
    [[ "$output" == *"$(own_arch)"* ]]
}

@test "a build for an empty arch fails and names the flag and the arch it builds for" {
    run "$MISO" build --arch=

    [ "$status" -eq 1 ]
    [[ "$output" == *'--arch ""'* ]]
    [[ "$output" == *"$(own_arch)"* ]]
}

@test "a build for the arch miso runs on is not refused for its arch" {
    # the cache of the test, which is not the user's
    export XDG_CACHE_HOME="$BATS_TEST_TMPDIR/cache"
    cd "$BATS_TEST_TMPDIR"

    run "$MISO" build --arch "$(own_arch)"

    # it fails later, on the missing Imagefile
    [[ "$output" != *"--arch"* ]]
}

# only a host that is not amd64 tells the two runs apart, see the arm64 job
@test "a build without --arch is a build for the arch miso runs on" {
    # the cache of the test, which is not the user's
    export XDG_CACHE_HOME="$BATS_TEST_TMPDIR/cache"
    cd "$BATS_TEST_TMPDIR"
    run "$MISO" build --arch "$(own_arch)"
    with_arch=$output

    run "$MISO" build

    [ "$output" == "$with_arch" ]
}

@test "the help of a build names the arch miso runs on as the default of --arch" {
    run "$MISO" build --help

    [ "$status" -eq 0 ]
    [[ "$output" == *"(default: \"$(own_arch)\")"* ]]
}
