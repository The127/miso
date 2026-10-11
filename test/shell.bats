#!/usr/bin/env bats
# MISO names the miso binary under test, see just test-cli.

@test "a shell for an arch it does not build for fails and names the flag and the arches it builds for" {
    run "$MISO" shell --arch sparc

    [ "$status" -eq 1 ]
    [[ "$output" == *'--arch "sparc"'* ]]
    [[ "$output" == *"amd64 and arm64"* ]]
}
