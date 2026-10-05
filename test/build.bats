#!/usr/bin/env bats
# MISO names the miso binary under test, see just test-cli.

@test "a build for an arch it does not build for fails and names the flag and the arch it builds for" {
    run "$MISO" build --arch sparc

    [ "$status" -eq 1 ]
    [[ "$output" == *'--arch "sparc"'* ]]
    [[ "$output" == *"amd64"* ]]
}

@test "a build for an empty arch fails and names the flag and the arch it builds for" {
    run "$MISO" build --arch=

    [ "$status" -eq 1 ]
    [[ "$output" == *'--arch ""'* ]]
    [[ "$output" == *"amd64"* ]]
}
