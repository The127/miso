#!/usr/bin/env bats
# MISO names the miso binary under test, see just test-cli.

@test "a prune with no limit fails and names the flags" {
    run "$MISO" prune

    [ "$status" -eq 1 ]
    [[ "$output" == *"--older-than"* ]]
    [[ "$output" == *"--keep-storage"* ]]
}

@test "a prune with a size it cannot read fails and names the flag" {
    run "$MISO" prune --keep-storage lots

    [ "$status" -eq 1 ]
    [[ "$output" == *"--keep-storage"* ]]
    [[ "$output" == *"lots"* ]]
}
